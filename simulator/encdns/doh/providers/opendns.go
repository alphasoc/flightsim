package providers

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"

	"github.com/alphasoc/flightsim/simulator/encdns"
	"golang.org/x/net/dns/dnsmessage"
)

type OpenDNS struct {
	Provider
}

// NewOpenDNS returns an *OpenDNS wrapping a Provider ready to use for DoH queries.
func NewOpenDNS(ctx context.Context) *OpenDNS {
	p := OpenDNS{
		Provider{
			addr:     "doh.opendns.com:443",
			queryURL: "https://doh.opendns.com/dns-query",
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

// QueryTXT performs a DoH TXT lookup on OpenDNS and returns an *encdns.Response and an
// error.
func (p *OpenDNS) QueryTXT(ctx context.Context, domain string) (*encdns.Response, error) {
	// OpenDNS requires requests to be in DNS wire format.
	dnsReq, err := encdns.NewUDPRequest(domain, dnsmessage.TypeTXT)
	if err != nil {
		return nil, err
	}
	reqStr := fmt.Sprintf("%v?dns=%v", p.queryURL, base64.StdEncoding.EncodeToString(dnsReq))
	req, err := http.NewRequestWithContext(ctx, "GET", reqStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "application/dns-message")
	clientResp, err := p.client.Do(req)
	return &encdns.Response{U: clientResp}, err
}
