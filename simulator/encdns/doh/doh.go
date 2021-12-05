// Package doh provides DNS-over-HTTPS functionality.
package doh

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/alphasoc/flightsim/simulator/encdns/dns"
	"golang.org/x/net/dns/dnsmessage"
)

// Question section of DoH query (JSON).
type Question struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

// Answer section of DoH query (JSON).
type Answer struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

// Response to a DoH query (JSON).
type Response struct {
	Status   int        `json:"Status"`
	TC       bool       `json:"TC"`
	RD       bool       `json:"RD"`
	RA       bool       `json:"RA"`
	AD       bool       `json:"AD"`
	CD       bool       `json:"CD"`
	Question []Question `json:"Question"`
	Answer   []Answer   `json:"Answer"`
	Comment  string     `json:"Comment"`
}

type Resolver struct {
	ctx         context.Context
	addr        string
	queryURL    string
	queryParams []string
	bindIP      net.IP
	c           *http.Client
}

type JSONResolver struct {
	Resolver
}

type WireResolver struct {
	Resolver
	wireProto string
}

// setupClient prepares the underlying communication structures for DoH.
func setupClient(ctx context.Context, addr string, bindIP net.IP) *http.Client {
	d := net.Dialer{}
	if bindIP != nil {
		d.LocalAddr = &net.TCPAddr{IP: bindIP}
	}
	tr := &http.Transport{
		// DoH uses TCP.
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.DialContext(ctx, "tcp", addr)
		},
	}
	c := &http.Client{Transport: tr}
	return c
}

// NewJSONResolver returns a ready to use DoH Resolver for basic JSON lookups.
func NewJSONResolver(
	ctx context.Context,
	addr, queryURL string,
	queryParams []string,
	bindIP net.IP) *JSONResolver {
	r := JSONResolver{
		Resolver: Resolver{
			ctx:         ctx,
			addr:        addr,
			queryURL:    queryURL,
			queryParams: queryParams,
			bindIP:      bindIP,
		},
	}
	r.c = setupClient(ctx, addr, bindIP)
	return &r
}

// NewWireResolver returns a ready to use DoH Resolver for lookups using the DNS wire
// protocol.
func NewWireResolver(
	ctx context.Context,
	addr, queryURL string,
	queryParams []string,
	wireProto string,
	bindIP net.IP) *WireResolver {
	r := WireResolver{
		Resolver: Resolver{
			ctx:         ctx,
			addr:        addr,
			queryURL:    queryURL,
			queryParams: queryParams,
			bindIP:      bindIP,
		},
		wireProto: wireProto,
	}
	r.c = setupClient(ctx, addr, bindIP)
	return &r
}

// genReqURL creates a request URL to be used for TXT lookups.  It returns the URL
// and an error.
func (r *JSONResolver) genReqURL(host string, t dnsmessage.Type) (string, error) {
	lenQueryParams := len(r.queryParams)
	if lenQueryParams == 0 {
		return "", fmt.Errorf("unable to generate DoH request URL: no query parameters")
	}
	reqURL := r.queryURL + "?"
	for i := 0; i < lenQueryParams; i++ {
		switch r.queryParams[i] {
		case "name":
			reqURL += fmt.Sprintf("name=%v", host)
		case "type":
			// Strip "Type" from the type name (ie. TypeTXT -> TXT).
			typeString := t.String()[len("Type"):]
			reqURL += fmt.Sprintf("type=%v", typeString)
		default:
			return "", fmt.Errorf(
				"unable to generate DoH request URL: unknown query parameter: '%v'",
				r.queryParams[i])
		}
		// Append '&' if not the last param.
		if i < lenQueryParams-1 {
			reqURL += "&"
		}
	}
	return reqURL, nil
}

// genReqURL creates a request URL to be used for TXT lookups.  It returns the URL
// and an error.
func (r *WireResolver) genReqURL(host string, t dnsmessage.Type) (string, error) {
	lenQueryParams := len(r.queryParams)
	// 1 query param needed for generating a wire request URL.
	if lenQueryParams != 1 {
		if lenQueryParams == 0 {
			return "", fmt.Errorf("unable to generate DoH request URL: no query parameters")
		}
		return "", fmt.Errorf("unable to generate DoH request URL: too manyquery parameters")
	}
	qParam := r.queryParams[0]
	var dnsReq []byte
	var err error
	if r.wireProto == "tcp" {
		dnsReq, err = dns.NewTCPRequest(host, t)
	} else if r.wireProto == "udp" {
		dnsReq, err = dns.NewUDPRequest(host, t)
	} else {
		err = fmt.Errorf(
			"unable to generate DoH request URL: invalid wire protocol: '%v'",
			r.wireProto)
	}
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
			"%v?%v=%v",
			r.queryURL,
			qParam,
			base64.StdEncoding.EncodeToString(dnsReq)),
		nil
}

// decode the JSON response, returning a slice of strings (records) and an error.
func (r *JSONResolver) decode(httpResp *http.Response) ([]string, error) {
	var records []string
	var resp Response
	err := json.NewDecoder(httpResp.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	for _, ans := range resp.Answer {
		records = append(records, strings.Trim(ans.Data, "\""))
	}
	return records, nil
}

// decode the wire response, returning a slice of strings (records) and an error.
func (r *WireResolver) decode(httpResp *http.Response) ([]string, error) {
	var buf bytes.Buffer
	n, err := io.Copy(&buf, httpResp.Body)
	// IO timeouts may be encountered.  If we managed to read anything, try to decrypt.
	if err != nil && n == 0 {
		return nil, err
	}
	return dns.ParseTXTResponse(buf.Bytes())
}

// LookupTXT perfors a DoH TXT lookup of host, returning TXT records as a slice of strings
// and an error.
func (r *JSONResolver) LookupTXT(ctx context.Context, host string) ([]string, error) {
	reqURL, err := r.genReqURL(host, dnsmessage.TypeTXT)
	if err != nil {
		return nil, err
	}
	// Form the HTTP GET request against the computed reqURL.
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	// Some DoH requests (depending on the server/provider) need this header.  For those
	// that don't require it, having it present doesn't seem to cause any errors (thus far).
	req.Header.Add("accept", "application/dns-json")
	// Perform the request.  On error, resp.Body already closed.
	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Decode the response, extracting the TXT records.
	records, err := r.decode(resp)
	if err != nil {
		return nil, err
	}
	return records, nil
}

// LookupTXT perfors a DoH TXT lookup of host, returning TXT records as a slice of strings
// and an error.
func (r *WireResolver) LookupTXT(ctx context.Context, host string) ([]string, error) {
	reqURL, err := r.genReqURL(host, dnsmessage.TypeTXT)
	if err != nil {
		return nil, err
	}
	// Form the HTTP GET request against the compute reqURL.
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	// Add the wire format header.
	req.Header.Add("accept", "application/dns-message")
	// Perform the request.  On error, resp.Body already closed.
	resp, err := r.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// Decode the response, extracting TXT records.
	records, err := r.decode(resp)
	if err != nil {
		return nil, err
	}
	return records, nil
}
