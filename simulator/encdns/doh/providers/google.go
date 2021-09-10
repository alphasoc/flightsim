package providers

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/alphasoc/flightsim/simulator/encdns"
)

type Google struct {
	Provider
}

// NewGoogle returns a *Google wrapping a Provider ready to use for DoH queries.
func NewGoogle(ctx context.Context) *Google {
	p := Google{
		Provider{
			addr:     "dns.google:443",
			queryURL: "https://dns.google/resolve",
		},
	}
	d := net.Dialer{}
	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.DialContext(ctx, "tcp", p.addr)
		},
	}
	p.client = &http.Client{Transport: tr}
	return &p
}

// QueryTXT performs a DoH TXT lookup on Google and returns an *encdns.Response and an
// error.
func (p *Google) QueryTXT(ctx context.Context, domain string) (*encdns.Response, error) {
	reqStr := fmt.Sprintf("%v?name=%v&type=TXT", p.queryURL, domain)
	req, err := http.NewRequestWithContext(ctx, "GET", reqStr, nil)
	if err != nil {
		return nil, err
	}
	clientResp, err := p.client.Do(req)
	return &encdns.Response{U: clientResp}, err
}
