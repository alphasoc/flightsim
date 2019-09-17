package simulator

import (
	"context"
	"fmt"
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
		PreferGo: true,
		Dial:     d.DialContext,
	}

	for i := 0; i < 40; i++ {
		label := strings.ToLower(utils.RandString(30))
		_, _ = r.LookupTXT(ctx, fmt.Sprintf("%s.%s", label, host))
		// TODO: make sure we get response
	}

	return nil
}

// Hosts returns random generated hosts to alphasoc sandbox.
func (t *Tunnel) Hosts(scope string, size int) ([]string, error) {
	return []string{"sandbox.alphasoc.xyz"}, nil
}
