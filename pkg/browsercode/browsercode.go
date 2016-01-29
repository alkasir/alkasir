package browsercode

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	mrand "math/rand"
	"net"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/thomasf/lg"
)

// BrowserCode .
type BrowserCode struct {
	Key  string // authentication key, HEX encoded.
	Host string // the client is bound to this, localhost is assumed if not specified.
	Port string // the client is bound to this,
}

func (b *BrowserCode) Addr() string {
	return fmt.Sprintf("%s:%s", b.Host, b.Port)
}

func (b *BrowserCode) Encode() (string, error) {
	host := b.Host
	if host == "localhost" {
		host = ""
	}
	if host == "127.0.0.1" {
		host = ""
	}
	if b.Key == "" {
		return "", errors.New("no key")
	}
	return fmt.Sprintf("%s:%s:%s", b.Key, host, b.Port), nil
}

func (b *BrowserCode) CopyToClipboard() error {
	encoded, err := b.Encode()
	if err != nil {
		return err
	}
	err = clipboard.WriteAll(encoded)
	if err != nil {
		return err
	}
	return nil

}

func Decode(encoded string) (BrowserCode, error) {
	parts := strings.Split(encoded, ":")
	if len(parts) != 3 {
		return BrowserCode{}, fmt.Errorf("wrong amount of parts in %s", encoded)
	}
	key := parts[0]
	host := parts[1]
	if host == "" {
		host = "localhost"
	}
	port := parts[2]

	return BrowserCode{
		Key:  key,
		Host: host,
		Port: port,
	}, nil

}
func (b *BrowserCode) SetHostport(hostport string) error {
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return err
	}
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		return fmt.Errorf("port must be supplied")
	}
	b.Port = port
	b.Host = host
	return nil
}

func New(hostport string) (BrowserCode, error) {
	bc := BrowserCode{}
	err := bc.SetHostport(hostport)
	if err != nil {
		return BrowserCode{}, err
	}
	key := NewKey()
	bc.Key = key
	return bc, nil
}

func NewKey() string {
	b := generateRandomBytes(24)
	key := hex.EncodeToString(b)
	return key
}

// generateRandomBytes returns securely generated random bytes. It will return
// an error if the system's secure random number generator fails to function
// correctly, in which case the caller should not continue.
func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		lg.Warningln("could not generate secure random key, using non secure random instead")
		lg.Warningln(err)
		for i := 0; i < n; i++ {
			b[i] = byte(mrand.Intn(256))
		}
	}
	return b
}
