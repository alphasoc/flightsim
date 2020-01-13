package simulator

import (
	"context"
	"io"
	"net"
	"strings"

	"github.com/alphasoc/flightsim/utils"
)

type Simulator interface {
	Simulate(ctx context.Context, bind net.IP, host string) error
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
}

func (TCPConnectSimulator) Simulate(ctx context.Context, bind net.IP, dst string) error {
	d := &net.Dialer{}
	if bind != nil {
		d.LocalAddr = &net.TCPAddr{IP: bind}
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
}

func (DNSResolveSimulator) Simulate(ctx context.Context, bind net.IP, dst string) error {
	host, _, _ := net.SplitHostPort(dst)
	if host == "" {
		host = dst
	}

	d := &net.Dialer{}
	if bind != nil {
		d.LocalAddr = &net.UDPAddr{IP: bind}
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
