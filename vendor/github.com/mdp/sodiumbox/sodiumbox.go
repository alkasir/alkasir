package sodiumbox

import (
	"crypto/rand"
	"errors"
	"io"

	"github.com/dchest/blake2b"

	"golang.org/x/crypto/nacl/box"
)

func extractKey(slice []byte) *[32]byte {
	a := new([32]byte)
	copy(a[:], slice[0:32])
	return a
}

func boxSealNonce(ephemeralPk, publicKey *[32]byte) (*[24]byte, error) {
	nonce := new([24]byte)
	hashConfig := &blake2b.Config{Size: 24}
	hashFn, err := blake2b.New(hashConfig)
	if err != nil {
		return nil, errors.New("Failed to create blake2b hash function")
	}
	hashFn.Write(ephemeralPk[0:32])
	hashFn.Write(publicKey[0:32])
	nonceSum := hashFn.Sum(nil)
	copy(nonce[:], nonceSum[0:24])
	return nonce, nil
}

// Message struct where we do all the work
type Message struct {
	Content     []byte
	Box         []byte
	PublicKey   *[32]byte
	SecretKey   *[32]byte
	EphemeralSK *[32]byte
	EphemeralPK *[32]byte
	Nonce       *[24]byte
	RandReader  *io.Reader
}

// Seal - seal the Message
// compatible with libsodium
func (m *Message) Seal() error {
	// ephemeral_pk || box(m, recipient_pk, ephemeral_sk, nonce=blake2b(ephemeral_pk || recipient_pk))
	if m.RandReader == nil {
		m.RandReader = &rand.Reader
	}
	ephemeralPK, ephemeralSK, err := box.GenerateKey(*m.RandReader)
	m.EphemeralPK = ephemeralPK
	m.EphemeralSK = ephemeralSK
	if err != nil {
		return errors.New("Failed to create ephemeral key pair")
	}
	nonce, err := boxSealNonce(m.EphemeralPK, m.PublicKey)
	if err != nil {
		return errors.New("Failed to build nonce")
	}
	m.Nonce = nonce
	boxed := box.Seal(nil, []byte(m.Content), m.Nonce, m.PublicKey, m.EphemeralSK)
	output := make([]byte, len(boxed)+32)
	copy(output[0:32], m.EphemeralPK[0:32])
	copy(output[32:], boxed[:])
	m.Box = output
	return nil
}

// Open - seal the Message
func (m *Message) Open() error {
	// ephemeral_pk || box(m, recipient_pk, ephemeral_sk, nonce=blake2b(ephemeral_pk || recipient_pk))
	m.EphemeralPK = extractKey(m.Box)
	nonce, err := boxSealNonce(m.EphemeralPK, m.PublicKey)
	if err != nil {
		return errors.New("Failed to build nonce")
	}
	m.Nonce = nonce
	boxed := make([]byte, len(m.Box)-32)
	copy(boxed, m.Box[32:])
	result, ok := box.Open(nil, boxed, m.Nonce, m.EphemeralPK, m.SecretKey)
	if !ok {
		return errors.New("Failed to decrypt")
	}
	m.Content = result
	return nil
}

// Seal crypto_box_seal_open
func Seal(content []byte, publicKey *[32]byte) (*Message, error) {
	message := &Message{
		Content:   content,
		PublicKey: publicKey,
	}
	err := message.Seal()
	if err != nil {
		return nil, errors.New("Failed to decrypt")
	}
	return message, nil
}

// SealOpen = crypto_box_seal_open
func SealOpen(enc []byte, publicKey, secretKey *[32]byte) (*Message, error) {
	message := &Message{
		Box:       enc,
		PublicKey: publicKey,
		SecretKey: secretKey,
	}
	err := message.Open()
	if err != nil {
		return nil, errors.New("Failed to decrypt")
	}
	return message, err
}
