package shared_test

import (
	"log"
	"reflect"
	"testing"

	. "github.com/alkasir/alkasir/pkg/shared"
)

var (
	hosts1 = []string{"23c.com.com", "com.com.github", "google.se",
		"something.something", "test.testtest.www", "www.youtub"}
	hosts2 = []string{"23c.com.com", "google.se", "help.localdomain",
		"something.something", "test.testtest.www", "www.youtube"}
)

func TestRevisionDiffPatch(t *testing.T) {
	v1 := &Revision{Content: hosts1}
	v2 := &Revision{Content: hosts2}
	diff := v1.Diff(v2)
	patched := v1.Patch(diff)

	if !reflect.DeepEqual(patched.Content, v2.Content) {
		log.Println(v1.Content)
		log.Println(v2.Content)
		t.Fail()
	}
}
