package simulator

import (
	"fmt"
	"math/rand"

	simssh "github.com/alphasoc/flightsim/simulator/ssh"
	bytesize "github.com/inhies/go-bytesize"
)

// SSHExfil defines this simulation.  It's little more than a wrapper around SSHTransfer.
type SSHExfil struct {
	SSHTransfer
}

// NewSSHExfil creates a new SSH Exfiltration simulator.
func NewSSHExfil() *SSHExfil {
	return &SSHExfil{}
}

// defualtTargetHosts returns a default string slice of targets in the {HOST:IP} form.
// Random selection of a non-standard SSH port is performed.
func (s *SSHExfil) defaultTargetHosts() []string {
	// Ports to be used for ssh exfil detectability.
	ports := []string{"443", "465", "993", "995"}
	pos := rand.Intn(len(ports))
	return []string{fmt.Sprintf("ssh.sandbox-services.alphasoc.xyz:%v", ports[pos])}
}

// defaultSendSize returns a 200 bytesize.MB default.
func (s *SSHExfil) defaultSendSize() bytesize.ByteSize {
	return 200 * bytesize.MB
}

// Hosts sets the simulation send size, and extracts the destination hosts.  A slice of
// strings representing the destination hosts (IP:port) is returned along with an error.
func (s *SSHExfil) Hosts(scope string, size int) ([]string, error) {
	dstHosts, sendSize, err := simssh.ParseScope(scope, s.defaultTargetHosts(), s.defaultSendSize())
	if err != nil {
		return dstHosts, err
	}
	s.sendSize = sendSize
	return dstHosts, nil
}
