package shared

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/alkasir/alkasir/pkg/namesgenerator"
)

type Connection struct {
	Transport string `json:"t"`
	Secret    string `json:"s"`
	Addr      string `json:"a"`
	Disabled  bool   `json:"disabled"`
	Protected bool   `json:"protected"`
	ID        string `json:"-"` // Generated for tracking connection history across runs
}

func (c *Connection) DisplayName() string {
	return fmt.Sprintf("%s (%s)",
		namesgenerator.GetPronouncableName(c.ID), c.Transport)
}

// the current version of the shareable data format, it could be any two
// letters possibly accepting more than one per version to avoid detection.
const ShareableTansportConnectionVersion = "ai"

// EnsureID creates a uniqie connection configuration identifier based on hash
// sum and stable across runs.
func (c *Connection) EnsureID() error {
	if c.ID == "" {
		hs := sha256.New()
		hs.Write([]byte("ADDR"))
		hs.Write([]byte(c.Addr))
		hs.Write([]byte("TRANSPORT"))
		hs.Write([]byte(c.Transport))
		hs.Write([]byte("SECRET"))
		hs.Write([]byte(c.Secret))
		resultWriter := new(bytes.Buffer)
		encoder := base64.NewEncoder(base64.RawURLEncoding, resultWriter)
		encoder.Write(hs.Sum(nil))
		encoder.Close()
		c.ID = string(resultWriter.Bytes()[:])
	}
	return nil
}

// Encode encodes the transport as a single string for the purpose of letting a
// user paste it into his application configuraion.
func (t *Connection) Encode() (string, error) {
	resultWriter := new(bytes.Buffer)

	// Write version number first and verify that it wrote exactly two characters
	n, _ := resultWriter.WriteString(ShareableTansportConnectionVersion)
	if n != 2 {
		panic("invalid version lentgh")
	}

	data := struct {
		Transport string `json:"t"`
		Secret    string `json:"s"`
		Addr      string `json:"a"`
	}{
		Transport: t.Transport,
		Secret:    t.Secret,
		Addr:      t.Addr,
	}
	// JSON encode the data struct
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	// Base64 encode the JSON data
	encoder := base64.NewEncoder(base64.URLEncoding, resultWriter)
	encoder.Write(jsonBytes)
	encoder.Close()

	return string(resultWriter.Bytes()[:]), nil
}

func (c *Connection) Decode(s string) error {
	s = strings.TrimSpace(s)
	if len(s) < 3 {
		return errors.New("Too short format")
	}

	// Split version and data into two strings
	version := s[0:2]
	data := s[2:]

	if version != ShareableTansportConnectionVersion {
		return errors.New("Unknown format version")
	}

	// Decode the data part as base64
	jsonBytes, err := base64.URLEncoding.DecodeString(data)
	if err != nil {
		return errors.New("Could not b64 decode: " + err.Error())
	}

	// Decode the previous result as json
	var connection Connection
	err = json.Unmarshal(jsonBytes, &connection)
	if err != nil {
		return errors.New("Cannot read json")
	}
	*c = connection
	return nil
}

func DecodeConnection(s string) (Connection, error) {
	c := Connection{}
	err := c.Decode(s)
	if err != nil {
		return c, err
	}
	err = c.EnsureID()

	return c, err

}
