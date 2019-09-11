package simulator

import (
	"context"
	"net"
)

type Simulator interface {
	Simulate(ctx context.Context, extIP net.IP, host string) error
}

type HostSource interface {
	Hosts(size int) ([]string, error)
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
		d.LocalAddr = &net.UDPAddr{IP: bind}
	}

	conn, err := d.DialContext(ctx, "tcp", dst)
	if err != nil {
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
	return err
}
