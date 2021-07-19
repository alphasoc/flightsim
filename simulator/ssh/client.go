// Package ssh provides just enough SSH/SFTP functionality to perform random writes
// to an SFTP server.
package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/alphasoc/flightsim/simulator/ssh/fxp"
	"golang.org/x/crypto/ssh"
)

// SSH/SFTP client version.  We don't perform any version negotiation.
const ClientVer = 3

// SSH/SFTP requests require an ID.  Per
// https://datatracker.ietf.org/doc/html/draft-ietf-secsh-filexfer-13#section-4, they
// don't need to be unique.  Since we're not sending requests/data in parallel, an id of
// 1 is just fine.
const reqID = 1

// Client wraps SSH client and session structs.
type Client struct {
	sshClient *ssh.Client
	sess      *ssh.Session
	w         io.WriteCloser
	r         io.Reader
}

// NewClient initializes and returns a Client ready to be used for SSH/SFTP transfer along
// with an error.  Note that Teardown should be called when the Client is no longer needed.
func NewClient(ctx context.Context, user string, src net.IP, dst string, signer ssh.Signer) (*Client, error) {
	// ClientConfig will use pubkey auth, ignore the host key and apply a 5 second connection
	// timeout.
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		// TODO insecure ok?
		//HostKeyCallback: ssh.FixedHostKey(hostKey),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}
	// Create dialer with src address, dial and setup a new client.
	d := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: src},
		Timeout:   config.Timeout,
	}
	nConn, err := d.DialContext(ctx, "tcp", dst)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}
	if deadLine, ok := ctx.Deadline(); ok {
		nConn.SetDeadline(deadLine)
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(nConn, dst, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}
	sshClient := ssh.NewClient(sshConn, chans, reqs)
	// Prep SFTP session and setup read/write pipes.
	sess, err := sshClient.NewSession()
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("failed initializing SFTP session: %w", err)
	}
	sess.RequestSubsystem("sftp")
	w, err := sess.StdinPipe()
	if err != nil {
		sess.Close()
		sshClient.Close()
		return nil, fmt.Errorf("failed initializing SFTP IO: %w", err)
	}
	r, err := sess.StdoutPipe()
	if err != nil {
		sess.Close()
		sshClient.Close()
		return nil, fmt.Errorf("failed initializing SFTP IO: %w", err)
	}
	return &Client{sshClient, sess, w, r}, nil
}

// Teardown closes the underlying SSH client and session.
func (c *Client) Teardown() {
	// ssh.Client.Close should close underlying net.Conn.
	c.sshClient.Close()
	// ssh.Session.Close will close channel used as the read/write pipes.
	c.sess.Close()
}

// ReadResp reads server responses to client requests.  A structure satisfying the Packet
// interface is returned along with an error.  The returned structure will either match
// the expectedRespType, or will be nil with an error.  If expectedRespType is a status,
// the caller is expected to check if the status is carrying an error.
func (c *Client) ReadResp(expectedRespType uint8) (Packet, error) {
	resp, err := ReadPacket(c.r)
	if err != nil {
		return nil, err
	}
	// First byte carries the packet type code, bytes 1 onward carry the payload.
	respType := resp[0]
	respData := resp[1:]
	// Handle status packets first, regardless whether that's the expected response, as
	// they may signal an error.
	if respType == fxp.TypeCodeStatus {
		// Unmarshalling may fail.
		statusResp, err := StatusResp(respData)
		if err != nil {
			return nil, err
		}
		// If the caller was expecting a status response, do no further processing.  The
		// caller is responsible for determining if this is an error.
		if expectedRespType == respType {
			return statusResp, nil
		}
		// Otherwise, check to see if the status response carries an error, and return it
		// along with a nil Packet.  We could add some additional error prefix
		// to the message, but it's becoming crowded with little added informational value.
		if statusResp.ErrCode != fxp.SSH_FX_OK {
			return nil, fmt.Errorf("%v", statusResp.ErrMsg)
		}
		// Otherwise, this appears to be an invalid response.
		return nil, fmt.Errorf("unexpected response type")
	}
	// We received something other than a status response.  It better be what the caller
	// wants.
	if expectedRespType != respType {
		return nil, fmt.Errorf("unexpected response type")
	}
	switch respType {
	case fxp.TypeCodeVersion:
		return VersionResp(respData)
	case fxp.TypeCodeHandle:
		return OpenResp(respData)
	default:
		return nil, fmt.Errorf("unsupported response type")
	}
}

// SendInit sends an init request to the server, returning a Version and an error.
func (c *Client) SendInit() (*fxp.Version, error) {
	initPkt := fxp.Init{Version: ClientVer}
	rawInitPkt := MakeRawPacket(&initPkt)
	if _, err := c.w.Write(rawInitPkt.Marshal()); err != nil {
		return nil, fmt.Errorf("failed init: %w", err)
	}
	resp, err := c.ReadResp(fxp.TypeCodeVersion)
	if err != nil {
		return nil, fmt.Errorf("failed init: %w", err)
	}
	if versionResp, ok := resp.(*fxp.Version); ok {
		return versionResp, nil
	}
	return nil, fmt.Errorf("failed init: invalid response processed")
}

// SendOpen sends an open filename request to the server and returns a Handle and an error.
func (c *Client) SendOpen(filename string, flags int) (*fxp.Handle, error) {
	openPkt := fxp.Open{ID: reqID, Filename: filename, Flags: uint32(flags)}
	rawOpenPkt := MakeRawPacket(&openPkt)
	if _, err := c.w.Write(rawOpenPkt.Marshal()); err != nil {
		return nil, fmt.Errorf("failed open: %w", err)
	}
	resp, err := c.ReadResp(fxp.TypeCodeHandle)
	if err != nil {
		return nil, fmt.Errorf("failed open: %w", err)
	}
	if openResp, ok := resp.(*fxp.Handle); ok {
		return openResp, nil
	}
	return nil, fmt.Errorf("failed open: invalid response processed")
}

// SendWrite sends a write request to the server, asking to write data at offset to the
// specified handle.  The number of bytes to send, number of bytes sent, a Status and
// an error are returned.
func (c *Client) SendWrite(handle string, offset uint64, data []byte) (int, int, *fxp.Status, error) {
	writePkt := fxp.Write{ID: reqID, Handle: handle, Offset: offset, Data: string(data)}
	rawWritePkt := MakeRawPacket(&writePkt)
	rawWritePktBytes := rawWritePkt.Marshal()
	rawWritePktBytesLen := len(rawWritePktBytes)
	bytesSent, err := c.w.Write(rawWritePktBytes)
	if err != nil {
		return rawWritePktBytesLen, bytesSent, nil, fmt.Errorf("failed write: %w", err)
	}
	resp, err := c.ReadResp(fxp.TypeCodeStatus)
	if err != nil {
		return rawWritePktBytesLen, bytesSent, nil, fmt.Errorf("failed write: %w", err)
	}
	writeResp, ok := resp.(*fxp.Status)
	if !ok {
		return rawWritePktBytesLen, bytesSent, nil, fmt.Errorf("failed write: invalid response processed")
	}
	if writeResp.ErrCode != fxp.SSH_FX_OK {
		return rawWritePktBytesLen, bytesSent, nil, fmt.Errorf("failed write: %v:", writeResp.ErrMsg)
	}
	return rawWritePktBytesLen, bytesSent, writeResp, nil
}

// SendClose sends a close file/handle request to the server.  A Status and an error are
// returned.
func (c *Client) SendClose(handle string) (*fxp.Status, error) {
	closePkt := fxp.Close{ID: reqID, Handle: handle}
	rawClosePkt := MakeRawPacket(&closePkt)
	if _, err := c.w.Write(rawClosePkt.Marshal()); err != nil {
		return nil, fmt.Errorf("failed close: %w", err)
	}
	resp, err := c.ReadResp(fxp.TypeCodeStatus)
	if err != nil {
		return nil, fmt.Errorf("failed close: %w", err)
	}
	closeResp, ok := resp.(*fxp.Status)
	if !ok {
		return nil, fmt.Errorf("failed close: invalid response processed")
	}
	if closeResp.ErrCode != fxp.SSH_FX_OK {
		return nil, fmt.Errorf("failed close: %v:", closeResp.ErrMsg)
	}
	return closeResp, nil
}
