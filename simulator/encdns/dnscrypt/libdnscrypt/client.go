// Package libdnscrypt provides the main functionality behind DNSCrypt.   Please see
// the LICENSE file in the libdnscrypt directory for further information.
package libdnscrypt

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/alphasoc/flightsim/simulator/encdns/dns"
	dnsstamps "github.com/jedisct1/go-dnsstamps"
	"golang.org/x/net/dns/dnsmessage"
)

// Error constants.
const (
	// ErrEsVersion means that the cert contains unsupported es-version
	ErrEsVersion = "unsupported es-version"

	// ErrInvalidDNSStamp means an invalid DNS stamp
	ErrInvalidDNSStamp = "invalid DNS stamp"

	// ErrCertTooShort means that it failed to deserialize cert, too short
	ErrCertTooShort = "cert is too short"

	// ErrCertMagic means an invalid cert magic
	ErrCertMagic = "invalid cert magic"

	// ErrInvalidQuery means that it failed to decrypt a DNSCrypt query
	ErrInvalidQuery = "DNSCrypt query is invalid and cannot be decrypted"

	// ErrInvalidResponse means that it failed to decrypt a DNSCrypt response
	ErrInvalidResponse = "DNSCrypt response is invalid and cannot be decrypted"

	// ErrInvalidClientMagic means that client-magic does not match
	ErrInvalidClientMagic = "DNSCrypt query contains invalid client magic"

	// ErrInvalidPadding means that it failed to unpad a query
	ErrInvalidPadding = "invalid padding"

	// ErrQueryTooLarge means that the DNS query is larger than max allowed size
	ErrQueryTooLarge = "DNSCrypt query is too large"

	// ErrInvalidResolverMagic means that server-magic does not match
	ErrInvalidResolverMagic = "DNSCrypt response contains invalid resolver magic"

	// ErrInvalidDNSResponse is an invalid DNS reponse error.
	ErrInvalidDNSResponse = "invalid DNS response"
)

// Basic Client struct.  Wrapped by Resolvers.
type Client struct {
	Net string
}

// ResolverInfo contains DNSCrypt resolver information necessary for decryption/encryption.
type ResolverInfo struct {
	SecretKey [keySize]byte // Client short-term secret key
	PublicKey [keySize]byte // Client short-term public key

	ServerPublicKey ed25519.PublicKey // Resolver public key (this key is used to validate cert signature)
	ServerAddress   string            // Server IP address
	ProviderName    string            // Provider name

	ResolverCert *Cert         // Certificate info (obtained with the first unencrypted DNS request)
	SharedKey    [keySize]byte // Shared key that is to be used to encrypt/decrypt messages
}

// findCertMagic is a bit of a hack to find the beginning of the certificate.  Returns
// the start of the certificate magic, or -1 if not found.
func findCertMagic(b []byte) int {
	return bytes.Index(b, certMagic[0:])
}

// fetchCert loads DNSCrypt cert from the specified server.
func (c *Client) fetchCert(ctx context.Context, stamp dnsstamps.ServerStamp) (*Cert, error) {
	providerName := stamp.ProviderName
	if !strings.HasSuffix(providerName, ".") {
		providerName = providerName + "."
	}

	dnsReq, err := dns.NewUDPRequest(providerName, dnsmessage.TypeTXT)
	if err != nil {
		return nil, err
	}
	d := net.Dialer{}
	// ctx, cancelFn := context.WithTimeout(ctx, 500*time.Millisecond)
	// defer cancelFn()
	conn, err := d.DialContext(ctx, c.Net, stamp.ServerAddrStr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write(dnsReq)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 2048)
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	n, err := conn.Read(b)
	if err != nil {
		return nil, err
	}
	// Check certificate response rcode==0.
	certMsg := dnsmessage.Message{}
	if err := certMsg.Unpack(b[0:n]); err != nil || certMsg.RCode != dnsmessage.RCodeSuccess {
		return nil, errors.New(ErrInvalidDNSResponse)
	}
	certIdx := findCertMagic(b)
	if certIdx == -1 {
		return nil, errors.New(ErrCertMagic)
	}
	certStr := b[certIdx:]
	cert := &Cert{}
	err = cert.Deserialize(certStr)
	if err != nil {
		return nil, err
	}
	return cert, nil
}

// Dial dials the server specified by stampStr, returning a *ResolverInfo and an error.
func (c *Client) Dial(ctx context.Context, stampStr string) (*ResolverInfo, error) {
	stamp, err := dnsstamps.NewServerStampFromString(stampStr)
	if err != nil {
		return nil, err
	}
	if stamp.Proto != dnsstamps.StampProtoTypeDNSCrypt {
		return nil, errors.New(ErrInvalidDNSStamp)
	}
	resolverInfo := &ResolverInfo{}
	// Generate the secret/public pair.
	resolverInfo.SecretKey, resolverInfo.PublicKey = generateRandomKeyPair()
	// Set the provider properties.
	resolverInfo.ServerPublicKey = stamp.ServerPk
	resolverInfo.ServerAddress = stamp.ServerAddrStr
	resolverInfo.ProviderName = stamp.ProviderName
	cert, err := c.fetchCert(ctx, stamp)
	if err != nil {
		return nil, err
	}
	resolverInfo.ResolverCert = cert
	// Compute shared key that we'll use to encrypt/decrypt messages.
	sharedKey, err := computeSharedKey(cert.EsVersion, &resolverInfo.SecretKey, &cert.ResolverPk)
	if err != nil {
		return nil, err
	}
	resolverInfo.SharedKey = sharedKey
	return resolverInfo, nil
}

// Encrypt encrypts a DNS message using shared key from the resolver info.  It returns a
// []byte and an error.
func (c *Client) Encrypt(m []byte, resolverInfo *ResolverInfo) ([]byte, error) {
	q := EncryptedQuery{
		EsVersion:   resolverInfo.ResolverCert.EsVersion,
		ClientMagic: resolverInfo.ResolverCert.ClientMagic,
		ClientPk:    resolverInfo.PublicKey,
	}
	b, err := q.Encrypt(m, resolverInfo.SharedKey)
	if len(b) > MinMsgSize {
		return nil, errors.New(ErrQueryTooLarge)
	}
	return b, err
}

// decrypts decrypts a DNS message using a shared key from the resolver info.  It returns
// a []byte and an error.
func (c *Client) Decrypt(b []byte, resolverInfo *ResolverInfo) ([]byte, error) {
	dr := EncryptedResponse{
		EsVersion: resolverInfo.ResolverCert.EsVersion,
	}
	msg, err := dr.Decrypt(b, resolverInfo.SharedKey)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
