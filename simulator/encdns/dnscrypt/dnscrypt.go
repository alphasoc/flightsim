// Package dnscrypt provides a DNSCrypt Resolver for TXT lookups.
package dnscrypt

import (
	"bytes"
	"context"
	"io"
	"net"
	"time"

	"github.com/alphasoc/flightsim/simulator/encdns/dns"
	"github.com/alphasoc/flightsim/simulator/encdns/dnscrypt/libdnscrypt"
	"golang.org/x/net/dns/dnsmessage"
)

type Resolver struct {
	ctx    context.Context
	sdns   string
	bindIP net.IP
	c      *libdnscrypt.Client
	d      *net.Dialer
	ri     *libdnscrypt.ResolverInfo
}

// NewResolver returns a pointer to a DNScrypt Resolver.
func NewResolver(ctx context.Context, network, sdns string, bindIP net.IP) *Resolver {
	return &Resolver{
		ctx:    ctx,
		sdns:   sdns,
		bindIP: bindIP,
		c:      &libdnscrypt.Client{Net: network}}
}

// prepConnection grabs DNSCrypt resolver info and prepares the underlying connection.
func (r *Resolver) prepConnection() error {
	// Connection already prepped.
	if r.ri != nil {
		return nil
	}
	ri, err := r.c.Dial(r.ctx, r.sdns)
	if err != nil {
		return err
	}
	r.ri = ri
	d := net.Dialer{}
	if r.bindIP != nil {
		if r.c.Net == "udp" {
			d.LocalAddr = &net.UDPAddr{IP: r.bindIP}
		} else {
			d.LocalAddr = &net.TCPAddr{IP: r.bindIP}
		}
	}
	r.d = &d
	return nil
}

// LookupTXT performs a DNSCrypt TXT lookup of host, returning TXT records as a slice of
// strings and an error.
func (r *Resolver) LookupTXT(ctx context.Context, host string) ([]string, error) {
	// On an initial lookup, get resolver information and server certificate.  Also,
	// prepare the dialer.
	var err error
	err = r.prepConnection()
	if err != nil {
		return nil, err
	}
	// Dial the actual server address obtained in ResliverInfo.
	conn, err := r.d.DialContext(r.ctx, r.c.Net, r.ri.ServerAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	var dnsReq []byte
	if r.c.Net == "udp" {
		dnsReq, err = dns.NewUDPRequest(host, dnsmessage.TypeTXT)
	} else {
		dnsReq, err = dns.NewTCPRequest(host, dnsmessage.TypeTXT)
	}
	if err != nil {
		return nil, err
	}
	// Encrypt the DNS wire protocol packet, and send.
	encryptedDnsReq, err := r.c.Encrypt(dnsReq, r.ri)
	if err != nil {
		return nil, err
	}
	_, err = conn.Write(encryptedDnsReq)
	if err != nil {
		return nil, err
	}
	// Set read deadline based on ctx.
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetReadDeadline(deadline)
	} else {
		conn.SetReadDeadline(time.Time{})
	}
	// Read the response, decrypting, and extracting the TXT records.
	var buf bytes.Buffer
	n, err := io.Copy(&buf, conn)
	// IO timeouts may be encountered.  If we managed to read anything, try to decrypt.
	if err != nil && n == 0 {
		return nil, err
	}
	resp := make([]byte, buf.Len())
	resp, err = r.c.Decrypt(buf.Bytes(), r.ri)
	if err != nil {
		return nil, err
	}
	return dns.ParseTXTResponse(resp)
}
