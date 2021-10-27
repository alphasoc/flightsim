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
type Tunnel struct {
	bind BindAddr
}

// NewTunnel creates dns tunnel simulator.
func NewTunnel() *Tunnel {
	return &Tunnel{}
}

func (s *Tunnel) Init(bind BindAddr) error {
	s.bind = bind
	return nil
}

func (Tunnel) Cleanup() {
}

// Simulate lookups for txt records for give host.
func (s *Tunnel) Simulate(ctx context.Context, host string) error {
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

	for {
		// keep going until the passed context expires
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		label := strings.ToLower(utils.RandString(30))

		ctx, cancelFn := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancelFn()
		_, err := r.LookupTXT(ctx, fmt.Sprintf("%s.%s", label, host))

		// Ignore "no such host".  Will ignore timeouts as well.
		if err != nil && !isSoftError(err, "no such host") {
			return err
		}

		// wait until context expires so we don't flood
		<-ctx.Done()
	}
}

// Hosts returns random generated hosts to alphasoc sandbox.
func (t *Tunnel) Hosts(scope string, size int) ([]string, error) {
	return []string{"sandbox.alphasoc.xyz"}, nil
}
