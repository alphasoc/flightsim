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
	Simulator
	HostSource
}

type DefaultSimulator struct {
	ip  TCPConnectSimulator
	dns DNSResolveSimulator
}

func (s *DefaultSimulator) Simulate(ctx context.Context, bind net.IP, dst string) error {
	var sim Simulator

	if ip := net.ParseIP(dst); ip != nil {
		sim = s.ip
	} else {
		sim = s.dns
	}

	return sim.Simulate(ctx, bind, dst)
}

type TCPConnectSimulator struct {
}

func (TCPConnectSimulator) Simulate(ctx context.Context, bind net.IP, dst string) error {
	d := &net.Dialer{
		LocalAddr: &net.TCPAddr{IP: bind},
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
	d := &net.Dialer{
		LocalAddr: &net.UDPAddr{IP: bind},
	}
	r := &net.Resolver{
		Dial: d.DialContext,
	}
	_, err := r.LookupHost(ctx, dst)
	return err
}
