package nexus

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"github.com/hanjos/nexus/util"
)

// Artifact is a Maven coordinate to a single artifact, plus the repository
// where it came from.
type Artifact struct {
	GroupID      string // e.g. org.springframework
	ArtifactID   string // e.g. spring-core
	Version      string // e.g. 4.1.3.RELEASE
	Classifier   string // e.g. sources, javadoc, <the empty string>...
	Extension    string // e.g. jar
	RepositoryID string // e.g. releases
}

// String implements the fmt.Stringer interface, as per Maven docs
// (http://maven.apache.org/pom.html#Maven_Coordinates).
func (a Artifact) String() string {
	var parts = []string{a.GroupID, a.ArtifactID, a.Extension}

	if a.Classifier != "" {
		parts = append(parts, a.Classifier)
	}

	return strings.Join(append(parts, a.Version), ":") + "@" + a.RepositoryID
}

// used for the artifact set.
func (a *Artifact) hash() string {
	return a.GroupID + ":" + a.ArtifactID + ":" + a.Version + ":" +
		a.Extension + ":" + a.Classifier + "@" + a.RepositoryID
}

// a zero-byte placeholder. No point in wasting bytes unnecessarily :)
var empty struct{}

// since Go doesn't have a built-in set implementation, a make-shift one
// follows, using a map for the heavy duty. Artifact's hash method is used to
// distinguish between artifacts, since there's no Java-like Equals contract to follow.
type artifactSet struct {
	// piles up the artifacts
	data []*Artifact

	// the set behavior
	hashMap map[string]struct{}
}

// creates and initializes a new set of artifacts.
func newArtifactSet() *artifactSet {
	return &artifactSet{
		data:    []*Artifact{},
		hashMap: make(map[string]struct{}),
	}
}

// adds a bunch of artifacts to this set.
func (set *artifactSet) add(artifacts []*Artifact) {
	for _, artifact := range artifacts {
		hash := artifact.hash()
		_, contains := set.hashMap[hash]

		set.hashMap[hash] = empty
		if !contains {
			set.data = append(set.data, artifact)
		}
	}
}

// ArtifactInfo holds extra information about an artifact. There are no
// constructors; use nexus.InfoOf to fetch and build instances.
type ArtifactInfo struct {
	*Artifact

	Uploader    string
	Uploaded    time.Time
	LastChanged time.Time
	Sha1        string
	Size        util.ByteSize
	MimeType    string
	URL         string
}

// keep this private; nexus.InfoOf will fill the ArtifactInfo out
func newInfoFromArtifact(artifact *Artifact) *ArtifactInfo {
	return &ArtifactInfo{Artifact: artifact}
}

// String implements the fmt.Stringer interface.
func (info ArtifactInfo) String() string {
	return fmt.Sprintf("%v : [SHA1 %v, Mime-Type %v, %v]",
		info.Artifact, info.Sha1, info.MimeType, info.Size)
}

// UnmarshalXML implements the xml.Unmarshaler interface. The ArtifactInfo
// should already have a valid *Artifact pointer; the rest of the information
// will be extracted from the payload.
func (info *ArtifactInfo) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	if info == nil {
		return fmt.Errorf("Can't unmarshal to a nil *ArtifactInfo!")
	}

	if info.Artifact == nil {
		return fmt.Errorf("Can't unmarshal an *ArtifactInfo with a nil *Artifact!")
	}

	var payload struct {
		Data struct {
			MimeType     string `xml:"mimeType"`
			Uploader     string `xml:"uploader"`
			Uploaded     int64  `xml:"uploaded"`
			LastChanged  int64  `xml:"lastChanged"`
			Size         int64  `xml:"size"`
			Sha1Hash     string `xml:"sha1Hash"`
			Repositories []struct {
				RepositoryID string `xml:"repositoryId"`
				ArtifactURL  string `xml:"artifactUrl"`
			} `xml:"repositories>org.sonatype.nexus.rest.model.RepositoryUrlResource"`
		} `xml:"data"`
	}

	if err := d.DecodeElement(&payload, &start); err != nil {
		return err
	}

	// finding the URL of this artifact. We need to know from which repository
	// the artifact is, so this is why we need the *Artifact already filled in
	url := ""
	for _, repo := range payload.Data.Repositories {
		if repo.RepositoryID == info.Artifact.RepositoryID {
			url = repo.ArtifactURL
			break
		}
	}

	info.Uploader = payload.Data.Uploader
	info.Uploaded = time.Unix(payload.Data.Uploaded, 0)
	info.LastChanged = time.Unix(payload.Data.LastChanged, 0)
	info.Sha1 = payload.Data.Sha1Hash
	info.Size = util.ByteSize(payload.Data.Size)
	info.MimeType = payload.Data.MimeType
	info.URL = url

	return nil
}

// A make-shift map-reducer, distributes an artifact search in multiple
// goroutines. Expects an array of strings and a query function. There will be
// one goroutine for every element of data. Each goroutine will call query with
// its respective datum.
func concurrentArtifactSearch(data []string, query func(string) ([]*Artifact, error)) ([]*Artifact, error) {
	artifacts := make(chan []*Artifact)
	errors := make(chan error)

	// search for the artifacts in each element of data
	for _, datum := range data {
		go func(datum string) {
			a, err := query(datum)
			if err != nil {
				errors <- err
				return
			}

			artifacts <- a
		}(datum)
	}

	// pile 'em up
	result := newArtifactSet()
	for i := 0; i < len(data); i++ {
		select {
		case a := <-artifacts:
			result.add(a)
		case err := <-errors:
			return nil, err
		}
	}

	return result.data, nil
}
