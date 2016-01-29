package search_test

import (
	"github.com/hanjos/nexus"
	"github.com/hanjos/nexus/credentials"
	"github.com/hanjos/nexus/search"
	"github.com/hanjos/nexus/util"

	"testing"
)

func TestAllImplementsCriteria(t *testing.T) {
	if _, ok := interface{}(search.All).(search.Criteria); !ok {
		t.Errorf("search.All does not implement Criteria!")
	}
}

func TestAllProvidesNoCriteria(t *testing.T) {
	criteria := search.All.Parameters()
	if len(criteria) != 0 {
		t.Errorf("expected an empty map, got %v", criteria)
	}
}

func TestByCoordinatesImplementsCriteria(t *testing.T) {
	if _, ok := interface{}(search.ByCoordinates{}).(search.Criteria); !ok {
		t.Errorf("search.ByCoordinates does not implement Criteria!")
	}
}

func TestByCoordinatesSetsTheProperFields(t *testing.T) {
	actual := search.ByCoordinates{GroupID: "g", ArtifactID: "a", Version: "v", Packaging: "p", Classifier: "c"}.Parameters()
	expected := map[string]string{"g": "g", "a": "a", "v": "v", "p": "p", "c": "c"}

	diff, onlyExpected, onlyActual := util.MapDiff(expected, actual)

	for _, key := range diff {
		t.Errorf("Mismatch on key %q: expected value %q, got %q", key, expected[key], actual[key])
	}

	for _, key := range onlyExpected {
		t.Errorf("Missing key %q", key)
	}

	for _, key := range onlyActual {
		t.Errorf("Unexpected key %q", key)
	}
}

func TestByClassnameImplementsCriteria(t *testing.T) {
	if _, ok := interface{}(search.ByClassname("")).(search.Criteria); !ok {
		t.Errorf("search.ByClassname does not implement Criteria!")
	}
}

func TestByClassnameSetsTheProperFields(t *testing.T) {
	actual := search.ByClassname("cn").Parameters()
	expected := map[string]string{"cn": "cn"}

	diff, onlyExpected, onlyActual := util.MapDiff(expected, actual)

	for _, key := range diff {
		t.Errorf("Mismatch on key %q: expected value %q, got %q", key, expected[key], actual[key])
	}

	for _, key := range onlyExpected {
		t.Errorf("Missing key %q", key)
	}

	for _, key := range onlyActual {
		t.Errorf("Unexpected key %q", key)
	}
}

func TestByChecksumImplementsCriteria(t *testing.T) {
	if _, ok := interface{}(search.ByChecksum("")).(search.Criteria); !ok {
		t.Errorf("search.ByChecksum does not implement Criteria!")
	}
}

func TestByChecksumSetsTheProperFields(t *testing.T) {
	actual := search.ByChecksum("sha1").Parameters()
	expected := map[string]string{"sha1": "sha1"}

	diff, onlyExpected, onlyActual := util.MapDiff(expected, actual)

	for _, key := range diff {
		t.Errorf("Mismatch on key %q: expected value %q, got %q", key, expected[key], actual[key])
	}

	for _, key := range onlyExpected {
		t.Errorf("Missing key %q", key)
	}

	for _, key := range onlyActual {
		t.Errorf("Unexpected key %q", key)
	}
}

func TestByKeywordImplementsCriteria(t *testing.T) {
	if _, ok := interface{}(search.ByKeyword("")).(search.Criteria); !ok {
		t.Errorf("search.ByKeyword does not implement Criteria!")
	}
}

func TestByKeywordSetsTheProperFields(t *testing.T) {
	actual := search.ByKeyword("q").Parameters()
	expected := map[string]string{"q": "q"}

	diff, onlyExpected, onlyActual := util.MapDiff(expected, actual)

	for _, key := range diff {
		t.Errorf("Mismatch on key %q: expected value %q, got %q", key, expected[key], actual[key])
	}

	for _, key := range onlyExpected {
		t.Errorf("Missing key %q", key)
	}

	for _, key := range onlyActual {
		t.Errorf("Unexpected key %q", key)
	}
}

func TestByRepositoryImplementsCriteria(t *testing.T) {
	if _, ok := interface{}(search.ByRepository("")).(search.Criteria); !ok {
		t.Errorf("search.ByRepository does not implement Criteria!")
	}
}

func TestByRepositorySetsTheProperFields(t *testing.T) {
	actual := search.ByRepository("repositoryId").Parameters()
	expected := map[string]string{"repositoryId": "repositoryId"}

	diff, onlyExpected, onlyActual := util.MapDiff(expected, actual)

	for _, key := range diff {
		t.Errorf("Mismatch on key %q: expected value %q, got %q", key, expected[key], actual[key])
	}

	for _, key := range onlyExpected {
		t.Errorf("Missing key %q", key)
	}

	for _, key := range onlyActual {
		t.Errorf("Unexpected key %q", key)
	}
}

func TestInRepositoryImplementsCriteria(t *testing.T) {
	if _, ok := interface{}(search.InRepository{}).(search.Criteria); !ok {
		t.Errorf("search.InRepository does not implement Criteria!")
	}
}

func TestInRepositorySetsTheProperFields(t *testing.T) {
	actual := search.InRepository{RepositoryID: "repositoryId", Criteria: search.ByChecksum("sha1")}.Parameters()
	expected := map[string]string{"repositoryId": "repositoryId", "sha1": "sha1"}

	diff, onlyExpected, onlyActual := util.MapDiff(expected, actual)

	for _, key := range diff {
		t.Errorf("Mismatch on key %q: expected value %q, got %q", key, expected[key], actual[key])
	}

	for _, key := range onlyExpected {
		t.Errorf("Missing key %q", key)
	}

	for _, key := range onlyActual {
		t.Errorf("Unexpected key %q", key)
	}
}

func TestInRepositoryWithSearchAllIsTheSameAsByRepository(t *testing.T) {
	actual := search.InRepository{RepositoryID: "repositoryId", Criteria: search.All}.Parameters()
	expected := search.ByRepository("repositoryId").Parameters()

	diff, onlyExpected, onlyActual := util.MapDiff(expected, actual)

	for _, key := range diff {
		t.Errorf("Mismatch on key %q: expected value %q, got %q", key, expected[key], actual[key])
	}

	for _, key := range onlyExpected {
		t.Errorf("Missing key %q", key)
	}

	for _, key := range onlyActual {
		t.Errorf("Unexpected key %q", key)
	}
}

// Examples

func ExampleByKeyword() {
	n := nexus.New("https://maven.java.net", credentials.None)

	// Return all artifacts with javax.enterprise somewhere.
	n.Artifacts(search.ByKeyword("javax.enterprise*"))

	// This search may or may not return an error, depending on the version of
	// the Nexus being accessed. On newer Nexuses (sp?) "*" searches are
	// invalid.
	n.Artifacts(search.ByKeyword("*"))
}

func ExampleByCoordinates() {
	n := nexus.New("https://maven.java.net", credentials.None)

	// Returns all artifacts with a groupID starting with com.sun. Due to Go's
	// struct syntax, we don't need to specify all the coordinates; they
	// default to string's zero value (""), which Nexus ignores.
	n.Artifacts(search.ByCoordinates{GroupID: "com.sun*"})

	// A coordinate search requires specifying at least either a groupID, an
	// artifactID or a version. This search will (after some time), return
	// nothing. This doesn't mean there are no projects with packaging "pom";
	// this is a limitation of Nexus' search.
	n.Artifacts(search.ByCoordinates{Packaging: "pom"})

	// This search may or may not return an error, depending on the version of
	// the Nexus being accessed. On newer Nexuses (sp?) "*" searches are
	// invalid.
	n.Artifacts(search.ByCoordinates{GroupID: "*", Packaging: "pom"})

	// ByCoordinates searches in Maven *projects*, not artifacts. So this
	// search will return all com.sun* artifacts in projects with packaging
	// "pom", not all POM artifacts with groupID com.sun*! Packaging is not
	// the same as extension.
	n.Artifacts(search.ByCoordinates{GroupID: "com*", Packaging: "pom"})
}

func ExampleInRepository() {
	n := nexus.New("https://maven.java.net", credentials.None)

	// Returns all artifacts in the repository releases with groupID starting
	// with com.sun and whose project has packaging "pom".
	n.Artifacts(
		search.InRepository{
			"releases",
			search.ByCoordinates{GroupID: "com.sun*", Packaging: "pom"},
		})

	// Nexus doesn't support * in the repository ID parameter, so this search
	// will return an error.
	n.Artifacts(
		search.InRepository{
			"releases*",
			search.ByCoordinates{GroupID: "com.sun*", Packaging: "pom"},
		})
}
