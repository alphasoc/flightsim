// Package dnscrypt provides DNSCrypt functionality.  It's pieced together from code found
// in the reference DNSCrypt implementation (https://github.com/DNSCrypt/dnscrypt-proxy),
// and in https://github.com/ameshkov/dnscrypt.git.  The goal was the provide FlightSim with
// just enough DNSCrypt, without pulling in too many non-golang.org third party libs.
package dnscrypt

import (
	"golang.org/x/net/dns/dnsmessage"
)

// IsValidResponse returns a boolean indicating if the *dnsmessage.Message carries a
// valid response.
func IsValidResponse(r *dnsmessage.Message) bool {
	return r.RCode == dnsmessage.RCodeSuccess && len(r.Answers) > 0
}

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
