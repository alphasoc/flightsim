package simulator

import (
	"context"
	"net"

	"github.com/pkg/errors"
)

// Hijack simulator.
type Hijack struct{}

// NewHijack creates port scan simulator.
func NewHijack() *Hijack {
	return &Hijack{}
}

// Simulate port scanning for given host.
func (*Hijack) Simulate(ctx context.Context, extIP net.IP, host string) error {
	d := &net.Dialer{
		LocalAddr: &net.UDPAddr{IP: extIP},
	}
	r := &net.Resolver{
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			return d.DialContext(ctx, "udp", "ns1.sandbox.alphasoc.xyz:53")
		},
	}

	addrs, err := r.LookupHost(ctx, host)
	if err != nil {
		return err
	}
	if len(addrs) > 1 {
		return errors.New("DNS domain hijacked")
	}
	return nil
}

// Hosts returns one domain to simulate dns query.
func (s *Hijack) Hosts(_ int) ([]string, error) {
	return []string{"alphasoc.com"}, nil
}
