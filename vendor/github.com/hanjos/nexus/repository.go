package nexus

import "encoding/xml"

// Repository is a non-group Nexus repository. Nexus actually provides a bit
// more data, but this should be enough for most uses. Groups aren't considered
// repositories by Nexus' API; there's a separate call for them.
type Repository struct {
	ID        string // e.g. releases
	Name      string // e.g. Releases
	Type      string // e.g. hosted, proxy, virtual...
	Format    string // e.g. maven2, maven1...
	Policy    string // e.g. RELEASE, SNAPSHOT
	RemoteURI string // e.g. http://repo1.maven.org/maven2/
}

// String implements the fmt.Stringer interface.
func (repo Repository) String() string {
	var uri string
	if repo.RemoteURI != "" {
		uri = ", points to " + repo.RemoteURI
	} else {
		uri = ""
	}

	return repo.ID + " ('" + repo.Name + "'){ " +
		repo.Type + ", " + repo.Format + " format, " +
		repo.Policy + " policy" + uri + " }"
}

// repos is here to help unmarshal Nexus' responses about repositories.
type repos []*Repository

// UnmarshalXML implements the xml.Unmarshaler interface.
func (r *repos) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var payload struct {
		Data []struct {
			ID        string `xml:"id"`
			Name      string `xml:"name"`
			Type      string `xml:"repoType"`
			Policy    string `xml:"repoPolicy"`
			Format    string `xml:"format"`
			RemoteURI string `xml:"remoteUri"`
		} `xml:"data>repositories-item"`
	}

	if err := d.DecodeElement(&payload, &start); err != nil {
		return err
	}

	for _, repo := range payload.Data {
		newRepo := &Repository{
			ID:        repo.ID,
			Name:      repo.Name,
			Type:      repo.Type,
			Format:    repo.Format,
			Policy:    repo.Policy,
			RemoteURI: repo.RemoteURI,
		}

		*r = append(*r, newRepo)
	}

	return nil
}
