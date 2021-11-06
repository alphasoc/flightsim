package providers

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/alphasoc/flightsim/simulator/encdns"
)

type CloudFlare struct {
	Provider
}

// NewCloudFlare returns a *CloudFlare wrapping a Provider ready to use for DoH queries.
func NewCloudFlare(ctx context.Context, bindIP net.IP) *CloudFlare {
	p := CloudFlare{
		Provider{
			addr:     "cloudflare-dns.com:443",
			queryURL: "https://cloudflare-dns.com/dns-query",
			bindIP:   bindIP,
		},
	}
	d := net.Dialer{}
	if bindIP != nil {
		d.LocalAddr = &net.TCPAddr{IP: p.bindIP}
	}
	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.DialContext(ctx, "tcp", p.addr)
		},
	}
	p.client = &http.Client{Transport: tr}
	return &p
}

// QueryTXT performs a DoH TXT lookup on CloudFlare and returns an *encdns.Response and an
// error.
func (p *CloudFlare) QueryTXT(ctx context.Context, domain string) (*encdns.Response, error) {
	reqStr := fmt.Sprintf("%v?name=%v&type=TXT", p.queryURL, domain)
	req, err := http.NewRequestWithContext(ctx, "GET", reqStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/dns-json")
	clientResp, err := p.client.Do(req)
	return &encdns.Response{U: clientResp}, err
}
