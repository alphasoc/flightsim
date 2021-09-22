// Package encdns provides general functionality for DoH, DoT and DNSCrypt.
package encdns

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

type Protocol int

// Supported DNS protocols.
const (
	Random Protocol = iota
	DoH
	DoT
	DNSCrypt
)

var protocolMap map[string]Protocol = map[string]Protocol{"doh": DoH, "dot": DoT, "dnscrypt": DNSCrypt}

// RandomProtocol returns a random supported Protocol.
func RandomProtocol() Protocol {
	// Account for the fact that Protocol(0) == Random.
	return Protocol(rand.Intn(len(protocolMap)) + 1)
}

// A generic response wrapper for DoH/DoT, etc.
type Response struct {
	U interface{}
}

// DOHResponse extracts a DoH (*http.Response) response and returns it along with an error.
func (r *Response) DOHResponse() (*http.Response, error) {
	httpResp, ok := r.U.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("not a DoH response")
	}
	return httpResp, nil
}

// DOTResponse extracts a DoT ([]string) response and returns it along with an error.
func (r *Response) DOTResponse() ([]string, error) {
	dotResp, ok := r.U.([]string)
	if !ok {
		return nil, fmt.Errorf("not a DoT response")
	}
	return dotResp, nil
}

// DNSCryptResponse extracts a DNSCrypt ([]byte) response, converts it to a
// *dnsmessage.Message and returns it along with an error.
func (r *Response) DNSCryptResponse() (*dnsmessage.Message, error) {
	dnsCryptResp, ok := r.U.([]byte)
	if !ok {
		return nil, fmt.Errorf("not a DNSCrypt response")
	}
	dnsMsg := &dnsmessage.Message{}
	if err := dnsMsg.Unpack(dnsCryptResp); err != nil {
		return nil, err
	}
	return dnsMsg, nil
}

// Queryable interface specifies the functionality that providers must implement.
type Queryable interface {
	QueryTXT(ctx context.Context, domain string) (*Response, error)
}

// NewTCPRequest creates a DNS wire request to be used over TCP.  It returns the request
// as a byte slice along with an error.
func NewTCPRequest(domain string, t dnsmessage.Type) ([]byte, error) {
	req, err := newRequest("tcp", domain, t)
	if err != nil {
		return nil, fmt.Errorf("failed creating DNS TCP request: %v", err)
	}
	return req, nil
}

// NewUDPRequest creates a DNS wire request to be used over UDP.  It returns the request
// as a byte slice along with an error.
func NewUDPRequest(domain string, t dnsmessage.Type) ([]byte, error) {
	req, err := newRequest("udp", domain, t)
	if err != nil {
		return nil, fmt.Errorf("failed creating DNS UDP request: %v", err)
	}
	return req[2:], nil
}

// newRequest creates a DNS wire request using the specified network protocol and returns
// the request as a byte slice along with an error
func newRequest(network string, domain string, t dnsmessage.Type) ([]byte, error) {
	name, err := dnsmessage.NewName(domain)
	if err != nil {
		return nil, err
	}
	q := dnsmessage.Question{
		Name:  name,
		Type:  t,
		Class: dnsmessage.ClassINET,
	}
	// TODO: dnsclient_unix.go::tryOneName() does nice error handling
	id := uint16(rand.Intn(256))
	b := dnsmessage.NewBuilder(
		make([]byte, 2, 514),
		dnsmessage.Header{ID: id, RecursionDesired: true})
	b.EnableCompression()
	if err := b.StartQuestions(); err != nil {
		return nil, err
	}
	if err := b.Question(q); err != nil {
		return nil, err
	}
	req, err := b.Finish()
	if err != nil {
		return nil, err
	}
	return req, err
}

// ParseScope parses the string scope and returns a Protocol and an error.
func ParseScope(scope string) (Protocol, error) {
	proto, ok := protocolMap[scope]
	// Invalid protocol requested; form an error message informing user of supported
	// protocols (DoH/DoT/etc).
	if !ok {
		protos := make([]string, len(protocolMap))
		i := 0
		for proto := range protocolMap {
			protos[i] = proto
			i++
		}
		scopes := fmt.Sprintf("%v", strings.Join(protos, ", "))
		return 0, fmt.Errorf("invalid commandline: invalid protocol '%v': protocol must be one of: %v", scope, scopes)
	}
	return proto, nil
}
