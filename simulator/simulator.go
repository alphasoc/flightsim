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

type Simulator interface {
	Init(bind net.IP) error
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
	bind net.IP
}

func (s *TCPConnectSimulator) Init(bind net.IP) error {
	s.bind = bind
	return nil
}

func (TCPConnectSimulator) Cleanup() {
}

func (s *TCPConnectSimulator) Simulate(ctx context.Context, dst string) error {
	d := &net.Dialer{}
	if s.bind != nil {
		d.LocalAddr = &net.TCPAddr{IP: s.bind}
	}

	conn, err := d.DialContext(ctx, "tcp", dst)
	if conn != nil {
		conn.Close()
	}

	if isSoftError(err, "connect: connection refused") {
		return nil
	}
	return err
}

type DNSResolveSimulator struct {
	bind net.IP
}

func (s *DNSResolveSimulator) Init(bind net.IP) error {
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
	if s.bind != nil {
		d.LocalAddr = &net.UDPAddr{IP: s.bind}
	}
	r := &net.Resolver{
		PreferGo: true,
		Dial:     d.DialContext,
	}
	_, err := r.LookupHost(ctx, utils.FQDN(host))

	if isSoftError(err, "no such host") {
		return nil
	}
	return err
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
