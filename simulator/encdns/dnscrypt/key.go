package dnscrypt

import (
	"crypto/rand"
	"errors"

	"github.com/aead/chacha20/chacha"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/nacl/box"
)

// generateRandomKeyPair generates a random key-pair.
func generateRandomKeyPair() (privateKey [keySize]byte, publicKey [keySize]byte) {
	privateKey = [keySize]byte{}
	publicKey = [keySize]byte{}

	_, _ = rand.Read(privateKey[:])
	curve25519.ScalarBaseMult(&publicKey, &privateKey)
	return
}

// computeSharedKey - computes a shared key.
func computeSharedKey(cryptoConstruction CryptoConstruction, secretKey *[keySize]byte, publicKey *[keySize]byte) ([keySize]byte, error) {
	if cryptoConstruction == XChacha20Poly1305 {
		sharedKey, err := sharedKey(*secretKey, *publicKey)
		if err != nil {
			return sharedKey, err
		}
		return sharedKey, nil
	} else if cryptoConstruction == XSalsa20Poly1305 {
		sharedKey := [sharedKeySize]byte{}
		box.Precompute(&sharedKey, publicKey, secretKey)
		return sharedKey, nil
	}
	return [keySize]byte{}, errors.New(ErrEsVersion)
}

// sharedKey computes a shared secret compatible with the one used by `crypto_box_xchacha20poly1305`.
func sharedKey(secretKey [32]byte, publicKey [32]byte) ([32]byte, error) {
	var sharedKey [32]byte

	sk, err := curve25519.X25519(secretKey[:], publicKey[:])
	if err != nil {
		return sharedKey, err
	}

	c := byte(0)
	for i := 0; i < 32; i++ {
		sharedKey[i] = sk[i]
		c |= sk[i]
	}
	if c == 0 {
		return sharedKey, errors.New("weak public key")
	}
	var nonce [16]byte
	chacha.HChaCha20(&sharedKey, &nonce, &sharedKey)
	return sharedKey, nil
}
