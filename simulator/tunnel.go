package simulator

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/utils"
)

// Tunnel simulator.
type Tunnel struct{}

// NewTunnel creates dns tunnel simulator.
func NewTunnel() *Tunnel {
	return &Tunnel{}
}

func (Tunnel) Init() error {
	return nil
}

func (Tunnel) Cleanup() {
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

	host = utils.FQDN(host)

	for {
		// keep going until the passed context expires
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		label := strings.ToLower(utils.RandString(30))

		ctx, _ := context.WithTimeout(ctx, 200*time.Millisecond)
		_, err := r.LookupTXT(ctx, fmt.Sprintf("%s.%s", label, host))

		// ignore timeout and "no such host"
		if err != nil && !isSoftError(err, "no such host") {
			return err
		}

		// wait until context expires so we don't flood
		<-ctx.Done()
	}

	return nil
}

// Hosts returns random generated hosts to alphasoc sandbox.
func (t *Tunnel) Hosts(scope string, size int) ([]string, error) {
	return []string{"sandbox.alphasoc.xyz"}, nil
}
