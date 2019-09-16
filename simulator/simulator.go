package simulator

import (
	"context"
	"net"
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
	if err != nil {
		if err, ok := err.(net.Error); ok {
			if err.Timeout() {
				return nil
			}
		}
		return err
	}
	conn.Close()

	return nil
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
		Dial: d.DialContext,
	}
	_, err := r.LookupHost(ctx, host)

	if err, ok := err.(*net.DNSError); ok {
		if err.IsNotFound || err.IsTimeout {
			return nil
		}
	}

	return err
}
