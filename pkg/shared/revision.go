package shared

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"time"

	"github.com/remyoudompheng/gigot/gitdelta"
)

// Revision represents a snapshot of the blocklist that if kept around can be
// used to send differential updates to users.
type Revision struct {
	Content []string  // The content of the
	Hash    string    // A hash of all the hostpatterns
	Created time.Time // When the list was downloaded
}

// NewRevision returns a Diff- and Patchable Revision based on it's content
func NewRevision(content []string) *Revision {
	rev := &Revision{
		Content: content,
		Created: time.Now(),
	}
	rev.makeHash()
	return rev
}

// InitialRevision returns an empty revision used as initial content
func InitialRevision() *Revision {
	rev := NewRevision(make([]string, 0))
	return rev
}

// MakeHash creates the Hash field based on the Content field
func (r *Revision) makeHash() {
	hasher := sha1.New()
	hasher.Write(r.bytes())
	r.Hash = hex.EncodeToString(hasher.Sum(nil))
}

// bytes returns the hostpatterns as a []byte for diffing and patching.
func (r *Revision) bytes() []byte {
	b := new(bytes.Buffer)
	for _, value := range r.Content {
		b.WriteString(value)
		b.WriteString("\n")
	}
	return b.Bytes()
}

// Diff returns a binary diff between two revisions.
func (r *Revision) Diff(newer *Revision) []byte {
	diff := gitdelta.Diff(r.bytes(), newer.bytes())
	return diff
}

// Patch applies a diff on a revision and returns a new revision patched to the
// latest version.
//
// TODO NOTE TODO NOTE This function will not be compleeted until the Revision
// struct is moved inside shared application code so that the client
// application can use it.
func (r *Revision) Patch(diff []byte) *Revision {
	oldbytes := r.bytes()
	patched, err := gitdelta.Patch(oldbytes, diff)
	if err != nil {
		panic(err)
	}
	content := make([]string, 0)
	reader := bytes.NewReader(patched)
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		item := scanner.Text()
		if item != "" {
			content = append(content, item)
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	return NewRevision(content)
}
