package simulator

import (
	"context"
	"io"
	"net"
	"strings"

	"github.com/alphasoc/flightsim/utils"
)

// HostMsgFormatter allows a simulator to implement a custom HostMsg method to be called in
// place of parsing the Module.HostMsg field.
type HostMsgFormatter interface {
	HostMsg(host string) string
}

// BindAddr wraps addr along with whether it was set by the user.  The UserSet flag is
// solely used by PipelineDNS simulators that specify their own dialer+LocalAddr.  In such
// cases, the usable/external IP is not always the correct choice (ie. on systems using
// systemd's stub resolver running on 127.0.0.53:53), so explicitly setting the LocalAddr
// is done only if the user has supplied an interface/address via flightsim's `-iface` flag.
type BindAddr struct {
	Addr    net.IP
	UserSet bool
}

// String returns the string representation of the underlying address.
func (b *BindAddr) String() string {
	return b.Addr.String()
}

type Simulator interface {
	Init(bind BindAddr) error
	Simulate(ctx context.Context, host string) error
	Cleanup()
}

// TODO: pass context
type HostSource interface {
	Hosts(scope string, size int) ([]string, error)
}

type Module interface {
	HostSource
	Simulator
}

func CreateModule(src HostSource, sim Simulator) Module {
	return &struct {
		HostSource
		Simulator
	}{src, sim}
}

type TCPConnectSimulator struct {
	bind BindAddr
}

func (s *TCPConnectSimulator) Init(bind BindAddr) error {
	s.bind = bind
	return nil
}

func (TCPConnectSimulator) Cleanup() {
}

func (s *TCPConnectSimulator) Simulate(ctx context.Context, dst string) error {
	d := &net.Dialer{LocalAddr: &net.TCPAddr{IP: s.bind.Addr}}

	conn, err := d.DialContext(ctx, "tcp", dst)
	if conn != nil {
		conn.Close()
	}
	// Ignore "connection refused" and timeouts.
	if err != nil && !isSoftError(err, "connect: connection refused") {
		return err
	}
	return nil
}

type DNSResolveSimulator struct {
	bind BindAddr
}

func (s *DNSResolveSimulator) Init(bind BindAddr) error {
	s.bind = bind
	return nil
}

func (DNSResolveSimulator) Cleanup() {
}

func (s *DNSResolveSimulator) Simulate(ctx context.Context, dst string) error {
	host, _, _ := net.SplitHostPort(dst)
	if host == "" {
		host = dst
	}

	d := &net.Dialer{}
	// Set the user overridden bind iface.
	if s.bind.UserSet {
		d.LocalAddr = &net.UDPAddr{IP: s.bind.Addr}
	}
	r := &net.Resolver{
		PreferGo: true,
		Dial:     d.DialContext,
	}

	host = utils.FQDN(host)

	_, err := r.LookupHost(ctx, host)
	// Ignore "no such host" and timeouts.
	if err != nil && !isSoftError(err, "no such host") {
		return err
	}

	return nil
}

func isTimeout(err error) bool {
	netErr, ok := err.(net.Error)
	if !ok {
		return false
	}
	return netErr.Timeout()
}

func isSoftError(err error, ss ...string) bool {
	if err == io.EOF {
		return true
	}
	netErr, ok := err.(net.Error)
	if !ok {
		return false
	}
	if netErr.Timeout() {
		return true
	}
	errStr := err.Error()
	for n := range ss {
		if strings.Contains(errStr, ss[n]) {
			return true
		}
	}
	return false
}
