package providers

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/alphasoc/flightsim/simulator/encdns"
)

type Quad9 struct {
	Provider
}

// NewQuad9 returns a *Quad9 wrapping a Provider ready to use for DoH queries.
func NewQuad9(ctx context.Context) *Quad9 {
	p := Quad9{
		Provider{
			addr:     "dns.quad9.net:5053",
			queryURL: "https://dns.quad9.net:5053/dns-query",
		},
	}
	d := net.Dialer{}
	tr := &http.Transport{
		DialContext: func(ctxt context.Context, network, addr string) (net.Conn, error) {
			return d.DialContext(ctx, "tcp", p.addr)
		},
	}
	p.client = &http.Client{Transport: tr}
	return &p
}

// QueryTXT performs a DoH TXT lookup on Quad9 and returns an *encdns.Response and
// an error.
func (p *Quad9) QueryTXT(ctx context.Context, domain string) (*encdns.Response, error) {
	reqStr := fmt.Sprintf("%v?name=%v&type=TXT", p.queryURL, domain)
	req, err := http.NewRequestWithContext(ctx, "GET", reqStr, nil)
	if err != nil {
		return nil, err
	}
	clientResp, err := p.client.Do(req)
	return &encdns.Response{U: clientResp}, err
}
