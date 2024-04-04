package simulator

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/lrstanley/girc"
)

// IRCDialer is a custom dialer for the girc library
// to use the given context in simulation
type IRCDialer struct {
	BindIP net.IP

	// context is used to set the deadline for the connection
	Ctx context.Context
}

// IRCDialer has to implement the girc.Dialer interface (Dial method)
func (ircd *IRCDialer) Dial(network, address string) (net.Conn, error) {
	d := net.Dialer{LocalAddr: &net.TCPAddr{IP: ircd.BindIP}}
	conn, err := d.DialContext(ircd.Ctx, network, address)

	if conn != nil {
		deadline, _ := ircd.Ctx.Deadline()
		err = conn.SetDeadline(deadline)
	}

	return conn, err
}

// IRCClient simulates IRC traffic
type IRCClient struct {
	bind BindAddr
}

// NewIRCClient creates new IRCClient simulator
func NewIRCClient() *IRCClient {
	return &IRCClient{}
}

func (irc *IRCClient) Init(bind BindAddr) error {
	irc.bind = bind
	return nil
}

func (IRCClient) Cleanup() {

}

// Simulate connection to IRC server
func (irc *IRCClient) Simulate(ctx context.Context, dst string) error {
	softErrors := []string{}

	// host is a domain name or IP address of the server
	// portStr is a string representation of the server port
	var host, portStr string

	// port is a port number of the server
	var port int
	var err error

	// If error occurs we assume it's a domain name without port
	if host, portStr, err = net.SplitHostPort(dst); err != nil {
		host = dst
		port = 6667
	} else {
		// Otherwise we assume it's an IP:port pair
		if port, err = strconv.Atoi(portStr); err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
	}

	// Create IRC client with given server address, port and credentials
	client := girc.New(girc.Config{
		Server:     host,
		Port:       port,
		Nick:       "nick" + randomHexID(5),
		User:       "user" + randomHexID(5),
		ServerPass: "password" + randomHexID(8),
	})

	// Disconnect on success
	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		client.Close()
	})

	dialer := &IRCDialer{Ctx: ctx, BindIP: irc.bind.Addr}

	if err = client.DialerConnect(dialer); err != nil {
		if isSoftError(err, softErrors...) {
			return nil
		}
		return err
	}

	return nil
}

// randomHexID generates a random hexstring of given lenght n
func randomHexID(n int) string {
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	buffer := make([]byte, n/2+1)

	_, _ = r.Read(buffer)
	return hex.EncodeToString(buffer)[:n]
}
