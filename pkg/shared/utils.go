package shared

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
)

// This type wraps a Write method and calls Sync after each Write.
type SyncWriter struct {
	*os.File
}

// Call File.Write and then Sync. An error is returned if either operation
// returns an error.
func (w SyncWriter) Write(p []byte) (n int, err error) {
	n, err = w.File.Write(p)
	if err != nil {
		return
	}
	err = w.Sync()
	return
}

// Generates atomic id seqences for low intensity rates.
// (There is probably something like this in the sync package already)
type idGen struct {
	mutex  sync.Mutex // modification lock
	count  uint64     // ticker for amount of generated id's
	prefix string     // entity prefix
}

// Return an atomic id generation instance, it will generate ids that looks
// like "prefix/n...".
func NewIDGen(prefix string) (*idGen, error) {
	if !validIDPrefix.MatchString(prefix) {
		return nil, errors.New("Invalid prefix")
	}

	generator := &idGen{
		mutex:  sync.Mutex{},
		count:  0,
		prefix: prefix,
	}
	return generator, nil
}

var validIDPrefix = regexp.MustCompile("^[[:alnum:]]+$")

// Returns the next prefix-(uint64) integer in the sequence from this instance.
func (i *idGen) New() (key string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.count = i.count + 1
	key = fmt.Sprintf("%s-%d", i.prefix, i.count)
	return
}

// generateRandomBytes returns securely generated random bytes. It will return
// an error if the system's secure random number generator fails to function
// correctly, in which case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString returns a hex encoded securely generated random string.
func SecureRandomString(s int) (string, error) {
	b, err := generateRandomBytes(s)
	return hex.EncodeToString(b), err
}

// SafeCleanReplace replaces all occurrences of a string in a string.
func SafeCleanReplace(string, replace string) string {
	return strings.Replace(string, replace, scrubbed, -1)
}

// SafeClean removes possibly sensitive information from a string which
// includes but might not be limited to ipv4 addresses.
func SafeClean(s string) string {
	return ipv4Matcher.ReplaceAllString(s, scrubbed)
}

var ipv4Matcher = regexp.MustCompile(
	ipv4AddrBlock + "\\." + ipv4AddrBlock + "\\." + ipv4AddrBlock + "\\." + ipv4AddrBlock)

const (
	ipv4AddrBlock = "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	scrubbed      = "[scrubbed]"
)
