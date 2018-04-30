package simulator

import (
	"context"
	"net"
	"strings"

	"github.com/alphasoc/flightsim/utils"
)

// Tunnel simulator.
type Tunnel struct{}

// NewTunnel creates dns tunnel simulator.
func NewTunnel() *Tunnel {
	return &Tunnel{}
}

// Simulate lookups for txt records for give host.
func (*Tunnel) Simulate(ctx context.Context, extIP net.IP, host string) error {
	d := &net.Dialer{
		LocalAddr: &net.UDPAddr{IP: extIP},
	}
	r := &net.Resolver{
		Dial: d.DialContext,
	}

	_, err := r.LookupTXT(ctx, host)
	return err
}

// Hosts returns random generated hosts to alphasoc sandbox.
func (t *Tunnel) Hosts(size int) ([]string, error) {
	var hosts []string

	for i := 0; i < size; i++ {
		label := strings.ToLower(utils.RandString(30))
		hosts = append(hosts, label+".sandbox.alphasoc.xyz")
	}

	return hosts, nil
}
