package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	simssh "github.com/alphasoc/flightsim/simulator/ssh"

	bytesize "github.com/inhies/go-bytesize"
)

const defaultSendSize = 100 * bytesize.MB

var defaultTargetHosts = []string{"sandbox.alphasoc.xyz:22", "sandbox.alphasoc.xyz:9999"}

// SSHTransfer defines this simulation.
type SSHTransfer struct {
	src       net.IP // Connect from this IP.
	sendSize  bytesize.ByteSize
	randomGen *rand.Rand
}

// NewSSHTransfer creates a new SSH/SFTP simulator.
func NewSSHTransfer() *SSHTransfer {
	return &SSHTransfer{}
}

// HostMsg implements the HostMsgFormatter interface, returning a custom host message
// string to be output by the run command.
func (s *SSHTransfer) HostMsg(host string) string {
	return fmt.Sprintf(
		"Simulating an SSH/SFTP file transfer of %v (%v) to %v",
		s.sendSize.Format("%.0f", "B", false),
		s.sendSize.Format("%.2f", "", false),
		host)
}

// Init sets the source IP for this simulation.
func (s *SSHTransfer) Init(src net.IP) error {
	s.src = src
	s.randomGen = rand.New(rand.NewSource(time.Now().UnixNano()))
	return nil
}

// writeRandom writes toSend bytes of 'random' data to the server.  The number of bytes
// sent and an error are returned.
func writeRandom(c *simssh.Client, handle string, toSend uint64, randomGen *rand.Rand) (uint64, error) {
	// 8K writes.
	const buffSize = 8192
	var totalDataBytesSent, leftToSend, i uint64
	var payloadSize int
	bytes := make([]byte, buffSize)
	for i = 0; totalDataBytesSent < toSend; i++ {
		leftToSend = toSend - totalDataBytesSent
		if leftToSend >= buffSize {
			payloadSize = buffSize
		} else {
			// Safe cast.
			payloadSize = int(leftToSend)
		}
		// Read always returns len(bytes) and a nil error.
		randomGen.Read(bytes[:payloadSize])
		// Care only about the number of bytes to send, number of bytes sent, and the err code.
		bytesToSend, bytesSent, _, err := c.SendWrite(handle, i*buffSize, bytes[:payloadSize])
		if err != nil {
			return totalDataBytesSent, fmt.Errorf("failed transfer: %w", err)
		}
		pktOverhead := bytesToSend - payloadSize
		dataBytesSent := bytesSent - pktOverhead
		if dataBytesSent > 0 {
			totalDataBytesSent += uint64(dataBytesSent)
		}
	}
	return totalDataBytesSent, nil
}

// Simulate an ssh/sftp file transfer.
func (s *SSHTransfer) Simulate(ctx context.Context, dst string) error {
	// Auth.
	signer, err := simssh.NewSignerFromKey()
	if err != nil {
		return err
	}
	// Create a Client that's ready to use for SSH/SFTP transfers.
	c, err := simssh.NewClient(ctx, "alphasoc", s.src, dst, signer)
	if err != nil {
		return err
	}
	defer c.Teardown()
	// Init/Version.
	initResp, err := c.SendInit()
	if err != nil {
		return err
	}
	// TODO: Do we really care about version mismatches?  From the sftp spec, a 3 can be
	// followed by some form of version negotiaion.
	if initResp.Version != simssh.ClientVer {
		return fmt.Errorf("server version mismatch, expecting %v, received %v",
			simssh.ClientVer, initResp.Version)
	}
	// Open a dummy file for writing and grab the handle returned by the server.  If used
	// with the sandbox SFTP server, no filesystem writes will actually occurr.
	openResp, err := c.SendOpen("flightsim-ssh-transfer", os.O_CREATE)
	if err != nil {
		return err
	}
	handle := openResp.Handle
	// Write s.sendSize bytes, checking for any write errors.
	bytesSent, err := writeRandom(c, handle, uint64(s.sendSize), s.randomGen)
	bytesizeBytesSent := bytesize.ByteSize(bytesSent)
	if err != nil {
		// Don't append ':" to leading '%v' as composed err already has trailing ':'.
		return fmt.Errorf(
			"%v [%v (%v) successfully transferred]",
			err,
			bytesizeBytesSent.Format("%.0f", "B", false),
			bytesizeBytesSent.Format("%.2f", "", false),
		)
	}
	// Close the handle.  We don't care about the response, just the error.
	_, err = c.SendClose(handle)
	if err != nil {
		return err
	}
	// Success.
	return nil
}

// Cleanup is a no-op.
func (s *SSHTransfer) Cleanup() {}

// parseScope parses the commandline portion (if supplied) after the module name.
//   ie. flightsim run ssh-transfer:this-part-is-scope:and-can-contain-more
// For the moment, only send size is parsed, but ultimately we also want to pass
// destination host and port.  Returns a string representation of the destination
// host (currently ""), the send size as a ByteSize, and an error.
func parseScope(scope string) ([]string, bytesize.ByteSize, error) {

	// scope can be "", in which case, apply defaults.
	if scope == "" {
		return defaultTargetHosts, defaultSendSize, nil
	}
	// scope may contain just the send size (ie. a lack of futher ":" separators
	// present in the string).
	var sendSize bytesize.ByteSize
	var err error
	if !strings.Contains(scope, ":") {
		sendSize, err = bytesize.Parse(scope)
		if err != nil {
			return []string{""},
				bytesize.ByteSize(0),
				fmt.Errorf("invalid command line: %w", err)
		}
		return defaultTargetHosts, sendSize, nil
	}
	// TODO scope may contain more information, separated by ":", perhaps as key-value
	// pairs.  For now, not supported.
	return []string{""}, bytesize.ByteSize(0), fmt.Errorf("invalid command line: %v", scope)
}

// Hosts sets the simulation send size, and extracts the destination hosts.  A slice of
// strings representing the destination hosts (IP:port) is returned along with an error.
func (s *SSHTransfer) Hosts(scope string, size int) ([]string, error) {
	dstHosts, sendSize, err := parseScope(scope)
	if err != nil {
		return dstHosts, err
	}
	s.sendSize = sendSize
	return dstHosts, nil
}
