// Package providers ...
// Additional providers can be found at: https://raw.githubusercontent.com/DNSCrypt/dnscrypt-resolvers/master/v2/public-resolvers.md
package providers

import (
	"context"
	"math/rand"
	"net"
	"time"

	"github.com/alphasoc/flightsim/simulator/encdns"
	"github.com/alphasoc/flightsim/simulator/encdns/dnscrypt"
	"golang.org/x/net/dns/dnsmessage"
)

type Provider struct {
	ctx       context.Context
	sdnsStamp string
	c         dnscrypt.Client
	bindIP    net.IP
}

// Providers supporting DNSCrypt.
var providers = []encdns.ProviderType{
	encdns.ScalewayFR,
	encdns.Yandex,
}

// NewRandom returns a 'random' Queryable provider.
func NewRandom(ctx context.Context, bindIP net.IP) encdns.Queryable {
	pIdx := encdns.ProviderType(rand.Intn(len(providers)))
	var p encdns.Queryable
	switch providers[pIdx] {
	case encdns.ScalewayFR:
		p = NewScalewayFR(ctx, bindIP)
	case encdns.Yandex:
		p = NewYandex(ctx, bindIP)
	}
	return p
}

// NewYandex returns a *Provider for Yandex's DNSCrypt service.
func NewYandex(ctx context.Context, bindIP net.IP) *Provider {
	return &Provider{
		ctx:       ctx,
		sdnsStamp: "sdns://AQQAAAAAAAAAEDc3Ljg4LjguNzg6MTUzNTMg04TAccn3RmKvKszVe13MlxTUB7atNgHhrtwG1W1JYyciMi5kbnNjcnlwdC1jZXJ0LmJyb3dzZXIueWFuZGV4Lm5ldA",
		c:         dnscrypt.Client{Net: "udp"},
		bindIP:    bindIP,
	}
}

// NewScalewayFR returns a *Provider for ScalewayFR's DNSCrypt service.
func NewScalewayFR(ctx context.Context, bindIP net.IP) *Provider {
	return &Provider{
		ctx:       ctx,
		sdnsStamp: "sdns://AQcAAAAAAAAADjIxMi40Ny4yMjguMTM2IOgBuE6mBr-wusDOQ0RbsV66ZLAvo8SqMa4QY2oHkDJNHzIuZG5zY3J5cHQtY2VydC5mci5kbnNjcnlwdC5vcmc",
		c:         dnscrypt.Client{Net: "udp"},
		bindIP:    bindIP,
	}
}

// QueryTXT performs a DNSCrypt TXT lookup using Provider p, and returns a *encdns.Response
// and an error.
func (p *Provider) QueryTXT(ctx context.Context, domain string) (*encdns.Response, error) {
	// Will obtain server cert, along with ResolverInfo for further queries.
	ri, err := p.c.Dial(p.ctx, p.sdnsStamp)
	if err != nil {
		return nil, err
	}
	d := net.Dialer{}
	if p.bindIP != nil {
		d.LocalAddr = &net.UDPAddr{IP: p.bindIP}
	}
	// Dial the actual server address obtained in ResliverInfo.
	conn, err := d.DialContext(p.ctx, p.c.Net, ri.ServerAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	dnsReq, err := encdns.NewUDPRequest(domain, dnsmessage.TypeTXT)
	if err != nil {
		return nil, err
	}
	// Encrypt the DNS UDP wire protocol packet, and send.
	encryptedDnsReq, err := p.c.Encrypt(dnsReq, ri)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(encryptedDnsReq)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 2048)
	// Set read deadline based on ctx.
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetReadDeadline(deadline)
	} else {
		conn.SetReadDeadline(time.Time{})
	}
	// Take note of bytes read for decrypt.
	n, err := conn.Read(b)
	if err != nil {
		return nil, err
	}
	r := make([]byte, 2048)
	r, err = p.c.Decrypt(b[0:n], ri)
	if err != nil {
		return nil, err
	}
	return &encdns.Response{U: r}, nil
}
