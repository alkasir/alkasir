package nexus

import (
	"encoding/xml"
	"testing"
)

func TestNexus2xImplementsClient(t *testing.T) {
	if _, ok := interface{}(Nexus2x{}).(Client); !ok {
		t.Errorf("nexus.Nexus2x does not implement nexus.Client!")
	}
}

func TestArtifactInfoPtrImplementsXmlUnmarshaler(t *testing.T) {
	if _, ok := interface{}(&ArtifactInfo{}).(xml.Unmarshaler); !ok {
		t.Errorf("nexus.ArtifactInfo does not implement xml.Unmarshaler!")
	}
}

var hasTests = []struct {
	input    map[string]string
	key      string
	expected bool
}{
	{
		map[string]string{"g": "g"},
		"g",
		true,
	},
	{
		map[string]string{"g": "g"},
		"a",
		false,
	},
	{
		map[string]string{"g": "g", "a": ""},
		"a",
		false,
	},
}

func TestHasWorks(t *testing.T) {
	for _, test := range hasTests {
		switch test.expected {
		case true:
			if _, ok := has(test.input, test.key); !ok {
				t.Errorf("Expected a value for key '%v' in %v, nothing found", test.key, test.input)
			}
		case false:
			if v, ok := has(test.input, test.key); ok {
				t.Errorf("Didn't expect anything for key '%v' in %v, found %v", test.key, test.input, v)
			}
		}
	}
}

func TestCantUnmarshalNilArtifactInfo(t *testing.T) {
	var info *ArtifactInfo

	err := info.UnmarshalXML(nil, xml.StartElement{})

	if err == nil {
		t.Errorf("Expected an error!")
		return
	}

	if err.Error() != "Can't unmarshal to a nil *ArtifactInfo!" {
		t.Errorf("Expected a different error, not '%v'", err.Error())
	}
}

func TestCantUnmarshalArtifactInfoWithANilArtifact(t *testing.T) {
	info := &ArtifactInfo{}

	err := info.UnmarshalXML(nil, xml.StartElement{})

	if err == nil {
		t.Errorf("Expected an error!")
		return
	}

	if err.Error() != "Can't unmarshal an *ArtifactInfo with a nil *Artifact!" {
		t.Errorf("Expected a different error, not '%v'", err.Error())
	}
}
