// Package dns provides some basic DNS utilities.
package dns

import (
	"fmt"
	"math/rand"

	"golang.org/x/net/dns/dnsmessage"
)

// NewTCPRequest creates a DNS wire request to be used over TCP.  It returns the request
// as a byte slice along with an error.
func NewTCPRequest(domain string, t dnsmessage.Type) ([]byte, error) {
	req, err := newRequest(domain, t)
	if err != nil {
		return nil, fmt.Errorf("failed creating DNS TCP request: %v", err)
	}
	lenReq := len(req) - 2
	req[0] = byte(lenReq >> 8)
	req[1] = byte(lenReq)
	return req, nil
}

// NewUDPRequest creates a DNS wire request to be used over UDP.  It returns the request
// as a byte slice along with an error.
func NewUDPRequest(domain string, t dnsmessage.Type) ([]byte, error) {
	req, err := newRequest(domain, t)
	if err != nil {
		return nil, fmt.Errorf("failed creating DNS UDP request: %v", err)
	}
	return req[2:], nil
}

// newRequest creates a DNS wire request using the specified network protocol and returns
// the request as a byte slice along with an error
func newRequest(domain string, t dnsmessage.Type) ([]byte, error) {
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

// ParseTXTResponse extracts TXT responses from a DNS message, and returns a slice of
// strings (the responses) and an error.
func ParseTXTResponse(msg []byte) ([]string, error) {
	p := dnsmessage.Parser{}
	_, err := p.Start(msg)
	if err != nil {
		return nil, err
	}
	err = p.SkipAllQuestions()
	if err != nil {
		return nil, err
	}
	var records []string
	// Loop over the answers, extracting TXT records, until ErrSectionDone.
	for {
		hdr, err := p.AnswerHeader()
		if err != nil {
			// Done.
			if err == dnsmessage.ErrSectionDone {
				break
			}
			// Unexpected error.
			return records, err
		}
		// Something other than a TXT record?
		if hdr.Type != dnsmessage.TypeTXT {
			return records, err
		}
		// Extract and append to the list of records.
		txt, err := p.TXTResource()
		if err == nil {
			records = append(records, txt.TXT...)
		}
	}
	return records, nil
}
