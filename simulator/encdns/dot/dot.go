// Package dot provides a DNS-over-TLS Resolver for TXT lookups.
package dot

import (
	"context"
	"crypto/tls"
	"net"
)

type Resolver struct {
	ctx    context.Context
	addr   string
	bindIP net.IP
	r      *net.Resolver
}

// NewResolver returns a ready to use, DoT Resolver.
func NewResolver(ctx context.Context, addr string, bindIP net.IP) *Resolver {
	r := Resolver{ctx: ctx, addr: addr, bindIP: bindIP}
	d := tls.Dialer{}
	if bindIP != nil {
		d.NetDialer = &net.Dialer{LocalAddr: &net.TCPAddr{IP: bindIP}}
	}
	r.r = &net.Resolver{
		PreferGo: true,
		// DoT uses TCP.
		Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.DialContext(r.ctx, "tcp", r.addr)
		},
	}
	return &r
}

// LookupTXT performs a DoT TXT lookup of host, returning TXT records as a slice of strings
// and an error.
func (r *Resolver) LookupTXT(ctx context.Context, host string) ([]string, error) {
	records, err := r.r.LookupTXT(ctx, host)
	if err != nil {
		return records, err
	}
	return records, nil
}
