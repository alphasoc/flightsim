package simulator

import (
	"context"
	"fmt"
	"net"
	"strings"

	simssh "github.com/alphasoc/flightsim/simulator/ssh"
	"github.com/alphasoc/flightsim/utils"
	"golang.org/x/crypto/ssh"

	bytesize "github.com/inhies/go-bytesize"
)

// SSHTransfer defines this simulation.
type SSHTransfer struct {
	src      net.IP // Connect from this IP.
	sendSize bytesize.ByteSize
}

// Client connection results struct.
type clientConnRes struct {
	c   *simssh.Client
	err error
}

// NewSSHTransfer creates a new SSH/SFTP simulator.
func NewSSHTransfer() *SSHTransfer {
	return &SSHTransfer{}
}

// defaultSendSize returns a 200 bytesize.MB default.
func (s *SSHTransfer) defaultSendSize() bytesize.ByteSize {
	return 200 * bytesize.MB
}

// defualtTargetHosts returns a default string slice of targets in the {HOST:IP} form.
func (s *SSHTransfer) defaultTargetHosts() []string {
	return []string{"ssh.sandbox-services.alphasoc.xyz:22"}
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
func (s *SSHTransfer) Init(src BindAddr) error {
	s.src = src.Addr
	return nil
}

// newClient initializes and returns SSH/SFTP Client along with an error.
func newClient(
	ctx context.Context,
	clientName string,
	src net.IP,
	dst string,
	signer ssh.Signer) (*simssh.Client, error) {
	// Create a Client that's ready to use for SSH/SFTP transfers.
	c, err := simssh.NewClient(ctx, clientName, src, dst, signer)
	if err != nil {
		// No need to invoke client Teardown(), as underlying connections would have been
		// closed by NewClient().
		return nil, err
	}
	// Init/Version.
	initResp, err := c.SendInit()
	if err != nil {
		c.Teardown()
		return c, err
	}
	// TODO: Do we really care about version mismatches?  From the sftp spec, a 3 can be
	// followed by some form of version negotiaion.
	if initResp.Version != simssh.ClientVer {
		c.Teardown()
		return c, fmt.Errorf("server version mismatch, expecting %v, received %v",
			simssh.ClientVer, initResp.Version)
	}
	return c, nil
}

type simulationContext struct {
	Ctx        context.Context
	Dst        string
	ClientName string
	Handle     string
	SendSize   bytesize.ByteSize
	Signer     ssh.Signer
	Ch         chan<- simssh.WriteResponse
}

// simulate performs the actual client connect and write on behalf of Simulate().
func (s *SSHTransfer) simulate(simCtx *simulationContext) {
	c, err := newClient(simCtx.Ctx, simCtx.ClientName, s.src, simCtx.Dst, simCtx.Signer)
	if err != nil {
		// Piggy back client connect errors on WriteResponse chan.
		res := simssh.WriteResponse{}
		res.ClientName = simCtx.ClientName
		res.Err = err
		simCtx.Ch <- res
		return
	}
	simCtx.Ch <- c.WriteRandom(simCtx.Handle, simCtx.SendSize)
	c.Teardown()
}

// Simulate an ssh/sftp file transfer.
func (s *SSHTransfer) Simulate(ctx context.Context, dst string) error {
	// Auth.
	signer, err := simssh.NewSignerFromKey()
	if err != nil {
		return err
	}
	// Compute number of clients and a send size for each, such that we don't exceed
	// maxClients.
	const maxClients = 2
	const minSendSize = 1 * bytesize.MB
	if s.sendSize <= 0 {
		return fmt.Errorf("invalid send size: %v", s.sendSize)
	}
	senderSizes := utils.ComputeSenderSizes(maxClients, s.sendSize, minSendSize)
	numClients := len(senderSizes)
	// Create a WriteResponse channel, used by clients to return connection errors and
	// write responses.
	writeCh := make(chan simssh.WriteResponse, numClients)
	for i := 0; i < numClients; i++ {
		go s.simulate(&simulationContext{
			Ctx:        ctx,
			Dst:        dst,
			ClientName: fmt.Sprintf("alphasoc-%v", i),
			Handle:     fmt.Sprintf("flightsim-ssh-transfer-%v", i),
			SendSize:   senderSizes[i],
			Signer:     signer,
			Ch:         writeCh})

	}
	var errsEncountered []string
	var totalBytesSent bytesize.ByteSize
	for i := 0; i < numClients; i++ {
		res := <-writeCh
		// Append all client connect and write errors, but continue.
		if res.Err != nil {
			errsEncountered = append(errsEncountered, fmt.Sprintf("client %v: %v", res.ClientName, res.Err.Error()))
		}
		totalBytesSent += bytesize.ByteSize(res.BytesSent)
	}
	if len(errsEncountered) != 0 {
		// Don't append ':" to leading '%v' as composed err already has trailing ':'.
		return fmt.Errorf(
			"[%v (%v) successfully transferred] Errors encountered:\n\t%v",
			totalBytesSent.Format("%.0f", "B", false),
			totalBytesSent.Format("%.2f", "", false),
			strings.Join(errsEncountered, "\n\t"),
		)
	}
	// Success.
	return nil
}

// Cleanup is a no-op.
func (s *SSHTransfer) Cleanup() {}

// Hosts sets the simulation send size, and extracts the destination hosts.  A slice of
// strings representing the destination hosts (IP:port) is returned along with an error.
func (s *SSHTransfer) Hosts(scope string, size int) ([]string, error) {
	dstHosts, sendSize, err := simssh.ParseScope(scope, s.defaultTargetHosts(), s.defaultSendSize())
	if err != nil {
		return dstHosts, err
	}
	s.sendSize = sendSize
	return dstHosts, nil
}
