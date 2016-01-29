package upgradebin

import (
	"crypto"
	"errors"
	"fmt"

	"github.com/agl/ed25519"
	"github.com/inconshreveable/go-update"
)

type verifyFn func([]byte, []byte, crypto.Hash, crypto.PublicKey) error

func (fn verifyFn) VerifySignature(checksum []byte, signature []byte, hash crypto.Hash, publicKey crypto.PublicKey) error {
	return fn(checksum, signature, hash, publicKey)
}

// NewECDSAVerifier returns a Verifier that uses the ECDSA algorithm to verify updates.
func NewED25519Verifier() update.Verifier {
	return verifyFn(func(checksum, signature []byte, hash crypto.Hash, publicKey crypto.PublicKey) error {
		key, ok := publicKey.(*[32]byte)
		if !ok {
			return errors.New("not a valid public key length")
		}

		if len(signature) != 64 {
			return fmt.Errorf("invalid signature length")
		}
		var sigRes = new([64]byte)
		for i, v := range signature {
			sigRes[i] = v
		}

		if !ed25519.Verify(key, checksum, sigRes) {
			return errors.New("failed to verify signature")
		}

		return nil
	})
}
