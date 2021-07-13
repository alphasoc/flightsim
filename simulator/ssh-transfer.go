package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	simssh "github.com/alphasoc/flightsim/simulator/ssh"

	bytesize "github.com/inhies/go-bytesize"
)

const defaultSendSize = 100 * bytesize.MB

// TODO where will the server side live?
var defaultTargetHosts = []string{"127.0.0.1:22", "127.0.0.1:9999"}

// SSHTransfer defines this simulation.
type SSHTransfer struct {
	src      net.IP // Connect from this IP.
	sendSize bytesize.ByteSize
}

// NewSSHTransfer creates a new SSH/SFTP simulator.
func NewSSHTransfer() *SSHTransfer {
	return &SSHTransfer{}
}

// Init sets the source IP for this simulation.
func (s *SSHTransfer) Init(src net.IP) error {
	s.src = src
	return nil
}

// writeRandom writes sendSize bytes of 'random' data to the server.  An error is returned.
func writeRandom(c *simssh.Client, handle string, sendSize bytesize.ByteSize) error {
	// 8K writes.
	const buffSize uint64 = 8192
	toSendStr := sendSize.Format("%.0f", "B", false)
	toSend, err := strconv.ParseUint(toSendStr[:len(toSendStr)-1], 10, 64)
	if err != nil {
		return fmt.Errorf("failed write: %v", err)
	}
	// Seed once.
	rand.Seed(time.Now().UnixNano())
	var i uint64
	var bytes []byte
	var payloadSize uint64
	for i = 0; toSend > 0; i++ {
		if toSend >= buffSize {
			payloadSize = buffSize
		} else {
			payloadSize = toSend
		}
		bytes = make([]byte, payloadSize)
		rand.Read(bytes)
		// Don't care about the actual response here.  Just the error code.
		_, err := c.SendWrite(handle, i*buffSize, &bytes)
		if err != nil {
			return fmt.Errorf("failed transfer: %v", err)
		}
		toSend -= payloadSize
	}
	return nil
}

// Simulate an ssh/sftp file transfer.
func (s *SSHTransfer) Simulate(ctx context.Context, dst string) error {
	// Auth.
	signer, err := simssh.NewSignerFromKey()
	if err != nil {
		return err
	}
	// Create a Client that's ready to use for SSH/SFTP transfers.
	c, err := simssh.NewClient("alphasoc", s.src, dst, signer)
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
	openResp, err := c.SendOpen("thereisnofile", os.O_CREATE)
	if err != nil {
		return err
	}
	handle := openResp.Handle
	// Write s.sendSize bytes, checking for any write errors.
	err = writeRandom(c, handle, s.sendSize)
	if err != nil {
		return err
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
				fmt.Errorf("invalid command line: %v", err)
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
