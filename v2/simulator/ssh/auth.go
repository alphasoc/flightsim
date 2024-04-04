package ssh

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// The private key here is solely used to authenticate with the server-side application
// accpeting the SSH/SFTP data transfer.  This is being done on purpose.
const privKey = `-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACA6El6xrgg3fzl6dbygysFLXmVN3ysbLnbnkD8jpgOAxQAAAJiL5q0ti+at
LQAAAAtzc2gtZWQyNTUxOQAAACA6El6xrgg3fzl6dbygysFLXmVN3ysbLnbnkD8jpgOAxQ
AAAED/EajTMrrGDzvT2VVeQhF/pf+mE9zINK7Kv3tHRynbpDoSXrGuCDd/OXp1vKDKwUte
ZU3fKxsudueQPyOmA4DFAAAAEGthcm9sQGVhc3kubG9jYWwBAgMEBQ==
-----END OPENSSH PRIVATE KEY-----`

// gSigner is the ssh signer used to authenticate.  It's global to this package to avoid
// computing it multiple times during a simulation
var gSigner ssh.Signer = nil

// NewSignerFromKey returns the global signer, if set, otherwise it generates an ssh signer
// from a static private key, sets the global signer varialbe, and returns the signer.  An
// error code is also returned.
func NewSignerFromKey() (ssh.Signer, error) {
	if gSigner != nil {
		return gSigner, nil
	}
	key := []byte(privKey)
	parsedKey, err := ssh.ParseRawPrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("unable to parse private key: %w", err)
	}
	signer, err := ssh.NewSignerFromKey(parsedKey)
	if err != nil {
		return nil, fmt.Errorf("unable to generate signer from private key: %w", err)
	}
	gSigner = signer
	return gSigner, nil
}
