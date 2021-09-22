package dnscrypt

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
)

// Cert is a DNSCrypt server certificate.
type Cert struct {
	// Serial is a 4 byte serial number in big-endian format. If more than
	// one certificates are valid, the client must prefer the certificate
	// with a higher serial number.
	Serial uint32

	// <es-version> ::= the cryptographic construction to use with this
	// certificate.
	// For X25519-XSalsa20Poly1305, <es-version> must be 0x00 0x01.
	// For X25519-XChacha20Poly1305, <es-version> must be 0x00 0x02.
	EsVersion CryptoConstruction

	// Signature is a 64-byte signature of (<resolver-pk> <client-magic>
	// <serial> <ts-start> <ts-end> <extensions>) using the Ed25519 algorithm and the
	// provider secret key. Ed25519 must be used in this version of the
	// protocol.
	Signature [ed25519.SignatureSize]byte

	// ResolverPk is the resolver's short-term public key, which is 32 bytes when using X25519.
	// This key is used to encrypt/decrypt DNS queries
	ResolverPk [keySize]byte

	// ResolverSk is the resolver's short-term private key, which is 32 bytes when using X25519.
	// Note that it's only used in the server implementation and never serialized/deserialized.
	// This key is used to encrypt/decrypt DNS queries
	ResolverSk [keySize]byte

	// ClientMagic is the first 8 bytes of a client query that is to be built
	// using the information from this certificate. It may be a truncated
	// public key. Two valid certificates cannot share the same <client-magic>.
	ClientMagic [clientMagicSize]byte

	// NotAfter is the date the certificate is valid from, as a big-endian
	// 4-byte unsigned Unix timestamp.
	NotBefore uint32

	// NotAfter is the date the certificate is valid until (inclusive), as a
	// big-endian 4-byte unsigned Unix timestamp.
	NotAfter uint32
}

// Deserialize deserializes certificate from a byte array
// <cert> ::= <cert-magic> <es-version> <protocol-minor-version> <signature>
//           <resolver-pk> <client-magic> <serial> <ts-start> <ts-end>
//           <extensions>
func (c *Cert) Deserialize(b []byte) error {
	if len(b) < 124 {
		return errors.New(ErrCertTooShort)
	}

	// <cert-magic>
	if !bytes.Equal(b[:4], certMagic[:4]) {
		return errors.New(ErrCertMagic)
	}

	// <es-version>
	switch esVersion := binary.BigEndian.Uint16(b[4:6]); esVersion {
	case uint16(XSalsa20Poly1305):
		c.EsVersion = XSalsa20Poly1305
	case uint16(XChacha20Poly1305):
		c.EsVersion = XChacha20Poly1305
	default:
		return errors.New(ErrEsVersion)
	}

	// Ignore 6:8, <protocol-minor-version>
	// <signature>
	copy(c.Signature[:], b[8:72])
	// <resolver-pk>
	copy(c.ResolverPk[:], b[72:104])
	// <client-magic>
	copy(c.ClientMagic[:], b[104:112])
	// <serial>
	c.Serial = binary.BigEndian.Uint32(b[112:116])
	// <ts-start> <ts-end>
	c.NotBefore = binary.BigEndian.Uint32(b[116:120])
	c.NotAfter = binary.BigEndian.Uint32(b[120:124])

	// Deserialized with no issues
	return nil
}
