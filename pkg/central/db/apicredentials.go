package db

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"errors"

	"golang.org/x/crypto/pbkdf2"
)

// APICredentials is used for export api jwt authentication.
type APICredentials struct {
	Username     string
	PasswordHash []byte
	Salt         []byte
	Enabled      bool
}

const passwordSaltLen = 16

func (a *APICredentials) IsValid(password string) (bool, error) {
	if !a.Enabled {
		return false, nil
	}
	if a.PasswordHash == nil {
		return false, errors.New("APICrendentials.PasswordHash is nil")
	}
	if a.Salt == nil {
		return false, errors.New("APICrendentials.Salt is nil")
	}
	if len(a.Salt) != passwordSaltLen {
		return false, errors.New("APICrendentials.Salt has wrong length")
	}
	dk := pbkdf2.Key([]byte(password), a.Salt, 4096, 32, sha1.New)
	return bytes.Equal(a.PasswordHash, dk), nil
}

func (a *APICredentials) SetPassword(password string) error {
	salt := make([]byte, passwordSaltLen)
	_, err := rand.Read(salt)
	if err != nil {
		return err
	}
	a.Salt = salt
	dk := pbkdf2.Key([]byte(password), a.Salt, 4096, 32, sha1.New)
	a.PasswordHash = dk
	return nil

}
