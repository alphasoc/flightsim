package libdnscrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"errors"
	"fmt"
	"time"
	"unsafe"

	"github.com/aead/chacha20/chacha"
	"golang.org/x/crypto/poly1305"
	"golang.org/x/crypto/salsa20/salsa"
)

// Crypto constants.
const (
	// See 11. Authenticated encryption and key exchange algorithm
	// The public and secret keys are 32 bytes long in storage
	keySize = 32

	// size of the shared key used to encrypt/decrypt messages
	sharedKeySize = 32

	// ClientMagic is the first 8 bytes of a client query that is to be built
	// using the information from this certificate. It may be a truncated
	// public key. Two valid certificates cannot share the same <client-magic>.
	clientMagicSize = 8

	// When using X25519-XSalsa20Poly1305, this construction requires a 24 bytes
	// nonce, that must not be reused for a given shared secret.
	nonceSize = 24

	// <min-query-len> is a variable length, initially set to 256 bytes, and
	// must be a multiple of 64 bytes. (see https://dnscrypt.info/protocol)
	// Some servers do not work if padded length is less than 256. Example: Quad9
	minUDPQuestionSize = 256

	// Minimum possible DNS packet size
	minDNSPacketSize = 12 + 5

	// the first 8 bytes of every dnscrypt response. must match resolverMagic.
	resolverMagicSize = 8
)

// Magic.
var (
	// certMagic is a bytes sequence that must be in the beginning of the serialized cert
	certMagic = [4]byte{0x44, 0x4e, 0x53, 0x43}

	// resolverMagic is a byte sequence that must be in the beginning of every response
	resolverMagic = []byte{0x72, 0x36, 0x66, 0x6e, 0x76, 0x57, 0x6a, 0x38}
)

// CryptoConstruction represents the encryption algorithm (either XSalsa20Poly1305 or XChacha20Poly1305).
type CryptoConstruction uint16

const (
	// UndefinedConstruction is the default value for empty CertInfo only
	UndefinedConstruction CryptoConstruction = iota
	// XSalsa20Poly1305 encryption
	XSalsa20Poly1305 CryptoConstruction = 0x0001
	// XChacha20Poly1305 encryption
	XChacha20Poly1305 CryptoConstruction = 0x0002
)

// Prior to encryption, queries are padded using the ISO/IEC 7816-4
// format. The padding starts with a byte valued 0x80 followed by a
// variable number of NUL bytes.
//
// ## Padding for client queries over UDP
//
// <client-query> <client-query-pad> must be at least <min-query-len>
// bytes. If the length of the client query is less than <min-query-len>,
// the padding length must be adjusted in order to satisfy this
// requirement.
//
// <min-query-len> is a variable length, initially set to 256 bytes, and
// must be a multiple of 64 bytes.
//
// ## Padding for client queries over TCP
//
// The length of <client-query-pad> is randomly chosen between 1 and 256
// bytes (including the leading 0x80), but the total length of <client-query>
// <client-query-pad> must be a multiple of 64 bytes.
//
// For example, an originally unpadded 56-bytes DNS query can be padded as:
//
// <56-bytes-query> 0x80 0x00 0x00 0x00 0x00 0x00 0x00 0x00
// or
// <56-bytes-query> 0x80 (0x00 * 71)
// or
// <56-bytes-query> 0x80 (0x00 * 135)
// or
// <56-bytes-query> 0x80 (0x00 * 199)
func pad(packet []byte) []byte {
	// get closest divisible by 64 to <packet-len> + 1 byte for 0x80
	minQuestionSize := (len(packet)+1+63)/64 + 64

	// padded size can't be less than minUDPQuestionSize
	minQuestionSize = max(minUDPQuestionSize, minQuestionSize)

	packet = append(packet, 0x80)
	for len(packet) < minQuestionSize {
		packet = append(packet, 0)
	}

	return packet
}

// unpad - removes padding bytes.
func unpad(packet []byte) ([]byte, error) {
	for i := len(packet); ; {
		if i == 0 {
			return nil, errors.New(ErrInvalidPadding)
		}
		i--
		if packet[i] == 0x80 {
			if i < minDNSPacketSize {
				return nil, errors.New(ErrInvalidPadding)
			}

			return packet[:i], nil
		} else if packet[i] != 0x00 {
			return nil, errors.New(ErrInvalidPadding)
		}
	}
}

const (
	// KeySize is what the name suggests
	KeySize = 32
	// NonceSize is what the name suggests
	NonceSize = 24
	// TagSize is what the name suggests
	TagSize = 16

	// MinMsgSize is the minimal size of a DNS packet.
	MinMsgSize = 512
)

func setup(subKey *[32]byte, counter *[16]byte, nonce *[24]byte, key *[32]byte) {
	// We use XSalsa20 for encryption so first we need to generate a
	// key and nonce with HSalsa20.
	var hNonce [16]byte
	copy(hNonce[:], nonce[:])
	salsa.HSalsa20(subKey, &hNonce, key, &salsa.Sigma)

	// The final 8 bytes of the original nonce form the new nonce.
	copy(counter[:], nonce[16:])
}

// sliceForAppend takes a slice and a requested number of bytes. It returns a
// slice with the contents of the given slice followed by that many bytes and a
// second slice that aliases into it and contains only the extra bytes. If the
// original slice has sufficient capacity then no allocation is performed.
func sliceForAppend(in []byte, n int) (head, tail []byte) {
	if total := len(in) + n; cap(in) >= total {
		head = in[:total]
	} else {
		head = make([]byte, total)
		copy(head, in)
	}
	tail = head[len(in):]
	return
}

// secretseal appends an encrypted and authenticated copy of message to out, which
// must not overlap message. The key and nonce pair must be unique for each
// distinct message and the output will be Overhead bytes longer than message.
func secretseal(out, message []byte, nonce *[24]byte, key *[32]byte) []byte {
	var subKey [32]byte
	var counter [16]byte
	setup(&subKey, &counter, nonce, key)

	// The Poly1305 key is generated by encrypting 32 bytes of zeros. Since
	// Salsa20 works with 64-byte blocks, we also generate 32 bytes of
	// keystream as a side effect.
	var firstBlock [64]byte
	salsa.XORKeyStream(firstBlock[:], firstBlock[:], &counter, &subKey)

	var poly1305Key [32]byte
	copy(poly1305Key[:], firstBlock[:])

	ret, out := sliceForAppend(out, len(message)+poly1305.TagSize)
	if AnyOverlap(out, message) {
		panic("nacl: invalid buffer overlap")
	}

	// We XOR up to 32 bytes of message with the keystream generated from
	// the first block.
	firstMessageBlock := message
	if len(firstMessageBlock) > 32 {
		firstMessageBlock = firstMessageBlock[:32]
	}

	tagOut := out
	out = out[poly1305.TagSize:]
	for i, x := range firstMessageBlock {
		out[i] = firstBlock[32+i] ^ x
	}
	message = message[len(firstMessageBlock):]
	ciphertext := out
	out = out[len(firstMessageBlock):]

	// Now encrypt the rest.
	counter[8] = 1
	salsa.XORKeyStream(out, message, &counter, &subKey)

	var tag [poly1305.TagSize]byte
	poly1305.Sum(&tag, ciphertext, &poly1305Key)
	copy(tagOut, tag[:])

	return ret
}

// Overhead is the number of bytes of overhead when boxing a message.
const Overhead = poly1305.TagSize

// secretopen authenticates and decrypts a box produced by Seal and appends the
// message to out, which must not overlap box. The output will be Overhead
// bytes smaller than box.
func secretopen(out, box []byte, nonce *[24]byte, key *[32]byte) ([]byte, bool) {
	if len(box) < Overhead {
		return nil, false
	}

	var subKey [32]byte
	var counter [16]byte
	setup(&subKey, &counter, nonce, key)

	// The Poly1305 key is generated by encrypting 32 bytes of zeros. Since
	// Salsa20 works with 64-byte blocks, we also generate 32 bytes of
	// keystream as a side effect.
	var firstBlock [64]byte
	salsa.XORKeyStream(firstBlock[:], firstBlock[:], &counter, &subKey)

	var poly1305Key [32]byte
	copy(poly1305Key[:], firstBlock[:])
	var tag [poly1305.TagSize]byte
	copy(tag[:], box)

	if !poly1305.Verify(&tag, box[poly1305.TagSize:], &poly1305Key) {
		return nil, false
	}

	ret, out := sliceForAppend(out, len(box)-Overhead)
	if AnyOverlap(out, box) {
		panic("nacl: invalid buffer overlap")
	}

	// We XOR up to 32 bytes of box with the keystream generated from
	// the first block.
	box = box[Overhead:]
	firstMessageBlock := box
	if len(firstMessageBlock) > 32 {
		firstMessageBlock = firstMessageBlock[:32]
	}
	for i, x := range firstMessageBlock {
		out[i] = firstBlock[32+i] ^ x
	}

	box = box[len(firstMessageBlock):]
	out = out[len(firstMessageBlock):]

	// Now decrypt the rest.
	counter[8] = 1
	salsa.XORKeyStream(out, box, &counter, &subKey)

	return ret, true
}

// seal does what the name suggests.
func seal(out, nonce, message, key []byte) []byte {
	if len(nonce) != NonceSize {
		panic("unsupported nonce size")
	}
	if len(key) != KeySize {
		panic("unsupported key size")
	}

	var firstBlock [64]byte
	cipher, _ := chacha.NewCipher(nonce, key, 20)
	cipher.XORKeyStream(firstBlock[:], firstBlock[:])
	var polyKey [32]byte
	copy(polyKey[:], firstBlock[:32])

	ret, out := sliceForAppend(out, TagSize+len(message))
	firstMessageBlock := message
	if len(firstMessageBlock) > 32 {
		firstMessageBlock = firstMessageBlock[:32]
	}

	tagOut := out
	out = out[poly1305.TagSize:]
	for i, x := range firstMessageBlock {
		out[i] = firstBlock[32+i] ^ x
	}
	message = message[len(firstMessageBlock):]
	ciphertext := out
	out = out[len(firstMessageBlock):]

	cipher.SetCounter(1)
	cipher.XORKeyStream(out, message)

	var tag [TagSize]byte
	hash := poly1305.New(&polyKey)
	_, _ = hash.Write(ciphertext)
	hash.Sum(tag[:0])
	copy(tagOut, tag[:])

	return ret
}

// open does what the name suggests.
func open(out, nonce, box, key []byte) ([]byte, error) {
	if len(nonce) != NonceSize {
		panic("unsupported nonce size")
	}
	if len(key) != KeySize {
		panic("unsupported key size")
	}
	if len(box) < TagSize {
		return nil, errors.New("ciphertext is too short")
	}

	var firstBlock [64]byte
	cipher, _ := chacha.NewCipher(nonce, key, 20)
	cipher.XORKeyStream(firstBlock[:], firstBlock[:])
	var polyKey [32]byte
	copy(polyKey[:], firstBlock[:32])

	var tag [TagSize]byte
	ciphertext := box[TagSize:]
	hash := poly1305.New(&polyKey)
	_, _ = hash.Write(ciphertext)
	hash.Sum(tag[:0])
	if subtle.ConstantTimeCompare(tag[:], box[:TagSize]) != 1 {
		return nil, errors.New("ciphertext authentication failed")
	}

	ret, out := sliceForAppend(out, len(ciphertext))

	firstMessageBlock := ciphertext
	if len(firstMessageBlock) > 32 {
		firstMessageBlock = firstMessageBlock[:32]
	}
	for i, x := range firstMessageBlock {
		out[i] = firstBlock[32+i] ^ x
	}
	ciphertext = ciphertext[len(firstMessageBlock):]
	out = out[len(firstMessageBlock):]

	cipher.SetCounter(1)
	cipher.XORKeyStream(out, ciphertext)
	return ret, nil
}

// EncryptedQuery is a structure for encrypting and decrypting client queries
//
// <dnscrypt-query> ::= <client-magic> <client-pk> <client-nonce> <encrypted-query>
// <encrypted-query> ::= AE(<shared-key> <client-nonce> <client-nonce-pad>, <client-query> <client-query-pad>)
type EncryptedQuery struct {
	// EsVersion is the encryption to use
	EsVersion CryptoConstruction

	// ClientMagic is a 8 byte identifier for the resolver certificate
	// chosen by the client.
	ClientMagic [clientMagicSize]byte

	// ClientPk is the client's public key
	ClientPk [keySize]byte

	// With a 24 bytes nonce, a question sent by a DNSCrypt client must be
	// encrypted using the shared secret, and a nonce constructed as follows:
	// 12 bytes chosen by the client followed by 12 NUL (0) bytes.
	//
	// The client's half of the nonce can include a timestamp in addition to a
	// counter or to random bytes, so that when a response is received, the
	// client can use this timestamp to immediately discard responses to
	// queries that have been sent too long ago, or dated in the future.
	Nonce [nonceSize]byte
}

// Encrypt encrypts the specified DNS query, returns encrypted data ready to be sent.
//
// Note that this method will generate a random nonce automatically.
//
// The following fields must be set before calling this method:
// * EsVersion -- to encrypt the query
// * ClientMagic -- to send it with the query
// * ClientPk -- to send it with the query
func (q *EncryptedQuery) Encrypt(packet []byte, sharedKey [sharedKeySize]byte) ([]byte, error) {
	var query []byte

	// Step 1: generate nonce
	binary.BigEndian.PutUint64(q.Nonce[:8], uint64(time.Now().UnixNano()))
	rand.Read(q.Nonce[8:12])

	// Unencrypted part of the query:
	// <client-magic> <client-pk> <client-nonce>
	query = append(query, q.ClientMagic[:]...)
	query = append(query, q.ClientPk[:]...)
	query = append(query, q.Nonce[:nonceSize/2]...)

	// <client-query> <client-query-pad>
	padded := pad(packet)

	// <encrypted-query>
	nonce := q.Nonce
	if q.EsVersion == XChacha20Poly1305 {
		query = seal(query, nonce[:], padded, sharedKey[:])
	} else if q.EsVersion == XSalsa20Poly1305 {
		var xsalsaNonce [nonceSize]byte
		copy(xsalsaNonce[:], nonce[:])
		query = secretseal(query, padded, &xsalsaNonce, &sharedKey)
	} else {
		return nil, errors.New(ErrEsVersion)
	}

	return query, nil
}

// Decrypt decrypts the client query, returns decrypted DNS packet.
//
// Please note, that before calling this method the following fields must be set:
// * ClientMagic -- to verify the query
// * EsVersion -- to decrypt
func (q *EncryptedQuery) Decrypt(query []byte, serverSecretKey [keySize]byte) ([]byte, error) {
	headerLength := clientMagicSize + keySize + nonceSize/2
	if len(query) < headerLength+TagSize+minDNSPacketSize {
		return nil, errors.New(ErrInvalidQuery)
	}

	// read and verify <client-magic>
	clientMagic := [clientMagicSize]byte{}
	copy(clientMagic[:], query[:clientMagicSize])
	if !bytes.Equal(clientMagic[:], q.ClientMagic[:]) {
		return nil, errors.New(ErrInvalidClientMagic)
	}

	// read <client-pk>
	idx := clientMagicSize
	copy(q.ClientPk[:keySize], query[idx:idx+keySize])

	// generate server shared key
	sharedKey, err := computeSharedKey(q.EsVersion, &serverSecretKey, &q.ClientPk)
	if err != nil {
		return nil, err
	}

	// read <client-nonce>
	idx = idx + keySize
	copy(q.Nonce[:nonceSize/2], query[idx:idx+nonceSize/2])

	// read and decrypt <encrypted-query>
	idx = idx + nonceSize/2
	encryptedQuery := query[idx:]
	var packet []byte
	if q.EsVersion == XChacha20Poly1305 {
		packet, err = open(nil, q.Nonce[:], encryptedQuery, sharedKey[:])
		if err != nil {
			return nil, errors.New(ErrInvalidQuery)
		}
	} else if q.EsVersion == XSalsa20Poly1305 {
		var xsalsaServerNonce [24]byte
		copy(xsalsaServerNonce[:], q.Nonce[:])
		var ok bool
		packet, ok = secretopen(nil, encryptedQuery, &xsalsaServerNonce, &sharedKey)
		if !ok {
			return nil, errors.New(ErrInvalidQuery)
		}
	} else {
		return nil, errors.New(ErrEsVersion)
	}

	packet, err = unpad(packet)
	if err != nil {
		return nil, errors.New(ErrInvalidPadding)
	}

	return packet, nil
}

// EncryptedResponse is a structure for encrypting/decrypting server responses
//
// <dnscrypt-response> ::= <resolver-magic> <nonce> <encrypted-response>
// <encrypted-response> ::= AE(<shared-key>, <nonce>, <resolver-response> <resolver-response-pad>)
type EncryptedResponse struct {
	// EsVersion is the encryption to use
	EsVersion CryptoConstruction

	// Nonce - <nonce> ::= <client-nonce> <resolver-nonce>
	// <client-nonce> ::= the nonce sent by the client in the related query.
	Nonce [nonceSize]byte
}

// Encrypt encrypts the server response
//
// EsVersion must be set.
// Nonce needs to be set to "client-nonce".
// This method will generate "resolver-nonce" and set it automatically.
func (r *EncryptedResponse) Encrypt(packet []byte, sharedKey [sharedKeySize]byte) ([]byte, error) {
	var response []byte

	// Step 1: generate nonce
	rand.Read(r.Nonce[12:16])
	binary.BigEndian.PutUint64(r.Nonce[16:nonceSize], uint64(time.Now().UnixNano()))

	// Unencrypted part of the query:
	response = append(response, resolverMagic[:]...)
	response = append(response, r.Nonce[:]...)

	// <resolver-response> <resolver-response-pad>
	padded := pad(packet)

	// <encrypted-response>
	nonce := r.Nonce
	if r.EsVersion == XChacha20Poly1305 {
		response = seal(response, nonce[:], padded, sharedKey[:])
	} else if r.EsVersion == XSalsa20Poly1305 {
		var xsalsaNonce [nonceSize]byte
		copy(xsalsaNonce[:], nonce[:])
		response = secretseal(response, padded, &xsalsaNonce, &sharedKey)
	} else {
		return nil, errors.New(ErrEsVersion)
	}

	return response, nil
}

// Decrypt decrypts the server response
//
// EsVersion must be set.
func (r *EncryptedResponse) Decrypt(response []byte, sharedKey [sharedKeySize]byte) ([]byte, error) {
	headerLength := len(resolverMagic) + nonceSize
	if len(response) < headerLength+TagSize+minDNSPacketSize {
		return nil, errors.New(ErrInvalidResponse)
	}

	// read and verify <resolver-magic>
	magic := [resolverMagicSize]byte{}
	copy(magic[:], response[:resolverMagicSize])
	if !bytes.Equal(magic[:], resolverMagic[:]) {
		return nil, errors.New(ErrInvalidResolverMagic)
	}

	// read nonce
	copy(r.Nonce[:], response[resolverMagicSize:nonceSize+resolverMagicSize])

	// read and decrypt <encrypted-response>
	encryptedResponse := response[nonceSize+resolverMagicSize:]
	var packet []byte
	var err error
	if r.EsVersion == XChacha20Poly1305 {
		packet, err = open(nil, r.Nonce[:], encryptedResponse, sharedKey[:])
		if err != nil {
			fmt.Println("FUCKED")

			return nil, errors.New(ErrInvalidResponse)
		}
	} else if r.EsVersion == XSalsa20Poly1305 {
		var xsalsaServerNonce [24]byte
		copy(xsalsaServerNonce[:], r.Nonce[:])
		var ok bool
		packet, ok = secretopen(nil, encryptedResponse, &xsalsaServerNonce, &sharedKey)
		if !ok {
			return nil, errors.New(ErrInvalidResponse)
		}
	} else {
		return nil, errors.New(ErrEsVersion)
	}

	packet, err = unpad(packet)
	if err != nil {
		return nil, errors.New(ErrInvalidPadding)
	}

	return packet, nil
}

// AnyOverlap reports whether x and y share memory at any (not necessarily
// corresponding) index. The memory beyond the slice length is ignored.
// Taken from the internal subtle package.
func AnyOverlap(x, y []byte) bool {
	return len(x) > 0 && len(y) > 0 &&
		uintptr(unsafe.Pointer(&x[0])) <= uintptr(unsafe.Pointer(&y[len(y)-1])) &&
		uintptr(unsafe.Pointer(&y[0])) <= uintptr(unsafe.Pointer(&x[len(x)-1]))
}
