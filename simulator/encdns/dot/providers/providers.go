// Package providers ...
package providers

import (
	"context"
	"crypto/tls"
	"math/rand"
	"net"

	"github.com/alphasoc/flightsim/simulator/encdns"
)

// Provider represents a DoT provider.  addr and ctx are used to dial.
type Provider struct {
	ctx    context.Context
	addr   string
	bindIP net.IP
}

// Providers supporting DoT.
var providers = []encdns.ProviderType{
	encdns.GoogleProvider,
	encdns.CloudFlareProvider,
	encdns.Quad9Provider,
	// OpenDNS does not, and does not plan to support DoT.
}

// NewRandom returns a 'random' Queryable provider.
func NewRandom(ctx context.Context, bindIP net.IP) encdns.Queryable {
	pIdx := encdns.ProviderType(rand.Intn(len(providers)))
	var p encdns.Queryable
	switch providers[pIdx] {
	case encdns.GoogleProvider:
		p = NewGoogle(ctx, bindIP)
	case encdns.CloudFlareProvider:
		p = NewCloudFlare(ctx, bindIP)
	case encdns.Quad9Provider:
		p = NewQuad9(ctx, bindIP)
	}
	return p
}

// NewGoogle returns a *Provider for Google's DoT service.
func NewGoogle(ctx context.Context, bindIP net.IP) *Provider {
	return &Provider{ctx: ctx, addr: "dns.google:853", bindIP: bindIP}
}

// NewGoogle returns a *Provider tied for CloudFlare's DoT service.
func NewCloudFlare(ctx context.Context, bindIP net.IP) *Provider {
	return &Provider{ctx: ctx, addr: "1dot1dot1dot1.cloudflare-dns.com:853", bindIP: bindIP}
}

// NewGoogle returns a *Provider for Quad9's DoT service.
func NewQuad9(ctx context.Context, bindIP net.IP) *Provider {
	return &Provider{ctx: ctx, addr: "dns.quad9.net:853", bindIP: bindIP}
}

// QueryTXT performs a DoT TXT lookup using Provider p, and returns a *encdns.Response and
// an error.
func (p *Provider) QueryTXT(ctx context.Context, domain string) (*encdns.Response, error) {
	d := tls.Dialer{}
	if p.bindIP != nil {
		d.NetDialer = &net.Dialer{LocalAddr: &net.TCPAddr{IP: p.bindIP}}
	}
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.DialContext(p.ctx, "tcp", p.addr)
		},
	}
	resp, err := r.LookupTXT(ctx, domain)
	return &encdns.Response{U: resp}, err
}
