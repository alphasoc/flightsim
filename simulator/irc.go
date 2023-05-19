package simulator

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"time"

	"github.com/lrstanley/girc"
)

// Custom dialer for the girc library
type IRCDialer struct {
	addr net.IP
	ctx  context.Context
}

// IrcDialer has to implement the girc.Dialer interface
func (ircd *IRCDialer) Dial(network, address string) (net.Conn, error) {
	d := net.Dialer{LocalAddr: &net.TCPAddr{IP: ircd.addr}}
	conn, err := d.DialContext(ircd.ctx, network, address)

	if conn != nil {
		deadline, _ := ircd.ctx.Deadline()
		err = conn.SetDeadline(deadline)
	}

	return conn, err
}

// IRC client simulator
type IRCClient struct {
	bind BindAddr
}

// NewIrcClient creates new IrcClient simulator
func NewIRCClient() *IRCClient {
	return &IRCClient{}
}

func (irc *IRCClient) Init(bind BindAddr) error {
	irc.bind = bind
	return nil
}

func (IRCClient) Cleanup() {

}

// Simulate connection to IRC server (PING - PONG)
func (irc *IRCClient) Simulate(ctx context.Context, dst string) error {
	/*
		TODO:
		consider potential soft errors:
		- connect: connection refused
		- read: connection reset by peer ?
		- You look like a bot ?
		- Password mismatch ?
		- no such host ?
		- timed out waiting for a requested PING response ?
		- i/o timeout ?

	*/

	softErrors := []string{"connect: connection refused"}

	client := girc.New(girc.Config{
		Server:     dst,
		Port:       6667,
		Nick:       "nick-" + randomId(),
		User:       "user-" + randomId(),
		ServerPass: "password-" + randomId(),
	})

	client.Handlers.Add(girc.CONNECTED, func(c *girc.Client, e girc.Event) {
		client.Close()
	})

	dialer := &IRCDialer{ctx: ctx, addr: irc.bind.Addr}

	if err := client.DialerConnect(dialer); err != nil {
		if isSoftError(err, softErrors...) {
			return nil
		}
		return err
	}

	return nil
}

func randomId() string {
	src := rand.NewSource(time.Now().Unix())
	r := rand.New(src)
	id := ""
	for i := 0; i < 8; i++ {
		id += strconv.Itoa(r.Intn(10))
	}

	return id
}
