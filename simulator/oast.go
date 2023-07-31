package simulator

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/utils"
)

// InteractshDefaultDomains is a list of default domains used by Interactsh.
var InteractshDefaultDomains = []string{
	"oast.fun",
	"oast.live",
	"oast.me",
	"oast.online",
	"oast.pro",
	"oast.site",
	"oastify.com",
}

// OAST simulator. This module simulates the out-of-band security testing (OAST) technique
// by trying to resolve random FQDNs under one of default domains used by Interactsh.
type OAST struct {
	bind BindAddr
}

// NewOAST creates OAST simulator.
func NewOAST() *OAST {
	return &OAST{}
}

func (oast *OAST) Init(bind BindAddr) error {
	oast.bind = bind
	return nil
}

func (OAST) Cleanup() {
}

// Simulate DNS lookups of random 33-character long hostnames beneath one of the default
// domains used by Interactsh.
func (oast *OAST) Simulate(ctx context.Context, host string) error {
	d := &net.Dialer{}
	// Set the user overridden bind iface.
	if oast.bind.UserSet {
		d.LocalAddr = &net.UDPAddr{IP: oast.bind.Addr}
	}
	r := &net.Resolver{
		PreferGo: true,
		Dial:     d.DialContext,
	}

	for {
		// Keep going until the passed context expires.
		select {
		case <-ctx.Done():
			return nil
		// Wait a random amount of time between 100ms and 500ms.
		case <-time.After(time.Duration(100+rand.Intn(400)) * time.Millisecond):
		}

		// Generate a random 33-character long hostname.
		hostname := strings.ToLower(utils.RandString(33))

		lctx, cancelFn := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancelFn()
		_, err := r.LookupIPAddr(lctx, fmt.Sprintf("%s.%s", hostname, host))

		// Ignore "no such host".  Will ignore timeouts as well.
		if err != nil && !isSoftError(err, "no such host") {
			return err
		}
	}
}

// Hosts returns a list of default domains used by Interactsh.
func (OAST) Hosts(scope string, size int) ([]string, error) {
	return InteractshDefaultDomains, nil
}
