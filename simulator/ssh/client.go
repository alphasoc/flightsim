// Package ssh provides just enough SSH/SFTP functionality to perform random writes
// to an SFTP server.
package ssh

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	sdipacket "github.com/alphasoc/sftp-data-input/packet"
	sdifxp "github.com/alphasoc/sftp-data-input/packet/fxp"

	"golang.org/x/crypto/ssh"
)

// SSH/SFTP client version.  We don't perform any version negotiation.
const ClientVer = 3

// Only SSH_FX_OK error code is needed for our purposes.
const SSH_FX_OK = 0

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
func (c *Client) ReadResp(expectedRespType uint8) (sdipacket.Packet, error) {
	resp, err := ReadPacket(c.r)
	if err != nil {
		return nil, err
	}
	// First byte carries the packet type code, bytes 1 onward carry the payload.
	respType := resp[0]
	respData := resp[1:]
	// Handle status packets first, regardless whether that's the expected response, as
	// they may signal an error.
	if respType == sdifxp.TypeCodeStatus {
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
		// along with a nil sdipacket.Packet.  We could add some additional error prefix
		// to the message, but it's becoming crowded with little added informational value.
		if statusResp.ErrCode != SSH_FX_OK {
			return nil, fmt.Errorf("%w", statusResp.ErrMsg)
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
	case sdifxp.TypeCodeVersion:
		return VersionResp(respData)
	case sdifxp.TypeCodeHandle:
		return OpenResp(respData)
	default:
		return nil, fmt.Errorf("unsupported response type")
	}
}

// SendInit sends an init request to the server, returning a Version and an error.
func (c *Client) SendInit() (*sdifxp.Version, error) {
	initPkt := sdifxp.Init{Version: ClientVer}
	rawInitPkt := sdipacket.MakeRawPacket(&initPkt)
	if _, err := c.w.Write(rawInitPkt.Marshal()); err != nil {
		return nil, fmt.Errorf("failed init: %w", err)
	}
	resp, err := c.ReadResp(sdifxp.TypeCodeVersion)
	if err != nil {
		return nil, fmt.Errorf("failed init: %w", err)
	}
	if versionResp, ok := resp.(*sdifxp.Version); ok {
		return versionResp, nil
	}
	return nil, fmt.Errorf("failed init: invalid response processed")
}

// SendOpen sends an open filename request to the server and returns a Handle and an error.
func (c *Client) SendOpen(filename string, flags int) (*sdifxp.Handle, error) {
	openPkt := sdifxp.Open{ID: reqID, Filename: filename, Flags: uint32(flags)}
	rawOpenPkt := sdipacket.MakeRawPacket(&openPkt)
	if _, err := c.w.Write(rawOpenPkt.Marshal()); err != nil {
		return nil, fmt.Errorf("failed open: %w", err)
	}
	resp, err := c.ReadResp(sdifxp.TypeCodeHandle)
	if err != nil {
		return nil, fmt.Errorf("failed open: %w", err)
	}
	if openResp, ok := resp.(*sdifxp.Handle); ok {
		return openResp, nil
	}
	return nil, fmt.Errorf("failed open: invalid response processed")
}

// SendWrite sends a write request to the server, asking to write data at offset to the
// specified handle.  A Status and an error are returned.
func (c *Client) SendWrite(handle string, offset uint64, data []byte) (*sdifxp.Status, error) {
	writePkt := sdifxp.Write{ID: reqID, Handle: handle, Offset: offset, Data: string(data)}
	rawWritePkt := sdipacket.MakeRawPacket(&writePkt)
	if _, err := c.w.Write(rawWritePkt.Marshal()); err != nil {
		return nil, fmt.Errorf("failed write: %w", err)
	}
	resp, err := c.ReadResp(sdifxp.TypeCodeStatus)
	if err != nil {
		return nil, fmt.Errorf("failed write: %w", err)
	}
	writeResp, ok := resp.(*sdifxp.Status)
	if !ok {
		return nil, fmt.Errorf("failed write: invalid response processed")
	}
	if writeResp.ErrCode != SSH_FX_OK {
		return nil, fmt.Errorf("failed write: %v:", writeResp.ErrMsg)
	}
	return writeResp, nil

}

// SendClose sends a close file/handle request to the server.  A Status and an error are
// returned.
func (c *Client) SendClose(handle string) (*sdifxp.Status, error) {
	closePkt := sdifxp.Close{ID: reqID, Handle: handle}
	rawClosePkt := sdipacket.MakeRawPacket(&closePkt)
	if _, err := c.w.Write(rawClosePkt.Marshal()); err != nil {
		return nil, fmt.Errorf("failed close: %w", err)
	}
	resp, err := c.ReadResp(sdifxp.TypeCodeStatus)
	if err != nil {
		return nil, fmt.Errorf("failed close: %w", err)
	}
	closeResp := resp.(*sdifxp.Status)
	if closeResp.ErrCode != SSH_FX_OK {
		return nil, fmt.Errorf("failed close: %v:", closeResp.ErrMsg)
	}
	return closeResp, nil
}
