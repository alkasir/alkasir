package upgradebin

import (
	"crypto"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"

	"github.com/agl/ed25519"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/inconshreveable/go-update"
	"github.com/thomasf/lg"
)

func NewUpdaterOptions(meta shared.BinaryUpgradeResponse, publicKey string) (update.Options, error) {
	sum, err := base64.RawURLEncoding.DecodeString(meta.SHA256Sum)
	if err != nil {
		return update.Options{}, err
	}

	sig, err := DecodeSignature(meta.ED25519Signature)
	if err != nil {
		return update.Options{}, err
	}
	pub, err := DecodePublicKey([]byte(publicKey))
	if err != nil {
		return update.Options{}, err
	}

	return update.Options{
		Patcher:   update.NewBSDiffPatcher(),
		Verifier:  NewED25519Verifier(meta.Version),
		Hash:      crypto.SHA256,
		Checksum:  sum,
		Signature: sig[:],
		PublicKey: pub,
	}, nil
}

// KeyPair .
type KeyPair struct {
	Public  *[32]byte
	Private *[64]byte
}

func DecodeKeys(priv, pub []byte) (*KeyPair, error) {
	pubK, err := DecodePublicKey(pub)
	if err != nil {
		return nil, err
	}
	privK, err := DecodePrivateKey(priv)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		Private: privK,
		Public:  pubK,
	}, nil
}

func DecodePrivateKey(priv []byte) (*[64]byte, error) {
	privblk, _ := pem.Decode([]byte(priv))
	if privblk == nil {
		return nil, fmt.Errorf("could not decode private key")
	}
	if privblk.Type != "ALKASIR UPGRADES PRIVATE KEY" {
		return nil, fmt.Errorf("invalid key type")
	}
	if len(privblk.Bytes) != 64 {
		return nil, fmt.Errorf("invalid length")
	}
	var privRes = new([64]byte)
	for i, v := range privblk.Bytes {
		privRes[i] = v
	}
	return privRes, nil
}

func DecodePublicKey(pub []byte) (*[32]byte, error) {
	pubblk, _ := pem.Decode([]byte(pub))
	if pubblk == nil {
		return nil, fmt.Errorf("could not decode public key")
	}
	if pubblk.Type != "ALKASIR UPGRADES PUBLIC KEY" {
		return nil, fmt.Errorf("invalid key type")
	}
	if len(pubblk.Bytes) != 32 {
		return nil, fmt.Errorf("invalid key length: %d", len(pubblk.Bytes))
	}

	var pubRes = new([32]byte)
	for i, v := range pubblk.Bytes {
		pubRes[i] = v
	}
	return pubRes, nil
}

func DecodeSignature(sig string) (*[64]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return nil, err
	}
	if len(data) != 64 {
		return nil, fmt.Errorf("invalid length")
	}
	var res = new([64]byte)
	for i, v := range data {
		res[i] = v
	}
	return res, nil
}

func EncodeKeys(privKey *[64]byte, pubKey *[32]byte) ([]byte, []byte) {
	priv := pem.EncodeToMemory(&pem.Block{
		Type:  "ALKASIR UPGRADES PRIVATE KEY",
		Bytes: privKey[:],
	})

	pub := pem.EncodeToMemory(&pem.Block{
		Type:  "ALKASIR UPGRADES PUBLIC KEY",
		Bytes: pubKey[:],
	})
	return priv, pub
}

func GenerateKeys(random io.Reader) (*[64]byte, *[32]byte) {
	pub, priv, err := ed25519.GenerateKey(random)
	if err != nil {
		lg.Fatal(err)
	}
	return priv, pub
}

func Apply(updateData io.Reader, opts update.Options) error {
	return update.Apply(updateData, opts)
}
