package util_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hanjos/nexus/util"
)

var cleanSlashesOK = []struct {
	input    string
	expected string
}{
	{"http://maven.java.net", "http://maven.java.net"},
	{"http://maven.java.net/", "http://maven.java.net/"},
	{"http://maven.java.net/////", "http://maven.java.net/"},
	{"http://maven.java.net/nexus", "http://maven.java.net/nexus"},
	{"http://maven.java.net/////nexus", "http://maven.java.net/nexus"},
}

var cleanSlashesErr = []struct {
	input    string
	expected reflect.Type
}{
	{"http:/maven.java.net", reflect.TypeOf(&util.MalformedURLError{})},
}

func TestCleanSlashes(t *testing.T) {
	for _, p := range cleanSlashesOK {
		actual, err := util.CleanSlashes(p.input)

		if err != nil {
			t.Errorf("expected %v, got an error %v", p.expected, err)
		} else if actual != p.expected {
			t.Errorf("expected %v, got %v", p.expected, actual)
		}
	}

	for _, p := range cleanSlashesErr {
		actual, err := util.CleanSlashes(p.input)

		if actual != "" {
			t.Errorf("expected an error %v, got a value %v", p.expected, actual)
		} else if reflect.TypeOf(err) != p.expected {
			t.Errorf("expected an error %v, got the error %v", p.expected, err)
		}
	}
}

type stringSet map[string]bool

func oneOf(args ...string) stringSet {
	set := map[string]bool{}

	for _, k := range args {
		set[k] = true
	}

	return set
}

var bfuOk = []struct {
	host  string
	path  string
	query map[string]string

	expected stringSet
}{
	{"http://maven.java.net", "nexus", map[string]string{},
		oneOf("http://maven.java.net/nexus")},
	{"http://maven.java.net", "///nexus", map[string]string{},
		oneOf("http://maven.java.net/nexus")},
	{"http://maven.java.net////", "/nexus", map[string]string{},
		oneOf("http://maven.java.net/nexus")},
	{"http://maven.java.net///", "/nexus", map[string]string{"p": "1"},
		oneOf("http://maven.java.net/nexus?p=1")},
	{"http://maven.java.net///", "/nexus", map[string]string{"p": "1", "q": "2"},
		oneOf("http://maven.java.net/nexus?p=1&q=2", "http://maven.java.net/nexus?q=2&p=1")},
}

var bfuErr = []struct {
	host  string
	path  string
	query map[string]string

	expected reflect.Type
}{
	{"http:/maven.java.net", "/nexus", map[string]string{}, reflect.TypeOf(&util.MalformedURLError{})},
}

func TestBuildFullURL(t *testing.T) {
	for _, p := range bfuOk {
		actual, err := util.BuildFullURL(p.host, p.path, p.query)

		if err != nil {
			t.Errorf("expected %v, got an error %v", p.expected, err)
		} else if _, ok := p.expected[actual]; !ok {
			t.Errorf("expected %v, got %v", p.expected, actual)
		}
	}

	for _, p := range bfuErr {
		actual, err := util.BuildFullURL(p.host, p.path, p.query)

		if actual != "" {
			t.Errorf("expected an error %v, got a value %v", p.expected, actual)
		} else if reflect.TypeOf(err) != p.expected {
			t.Errorf("expected an error %v, got an error %v", p.expected, err)
		}
	}
}

func TestIfMalformedURLErrorIsError(t *testing.T) {
	// type assertion only works on interface types, so...
	if _, ok := interface{}(&util.MalformedURLError{}).(error); !ok {
		t.Errorf("util.MalformedURLError does not implement the error interface!")
	}
}

var mdTest = []struct {
	expected map[string]string
	actual   map[string]string

	diff         []string
	onlyExpected []string
	onlyActual   []string
}{
	{
		map[string]string{"a": "a", "b": "b"},
		map[string]string{"a": "a1", "c": "c"},

		[]string{"a"},
		[]string{"b"},
		[]string{"c"},
	},
	{
		map[string]string{"a": "a", "b": "b"},
		map[string]string{"a": "a", "b": "b"},

		[]string{},
		[]string{},
		[]string{},
	},
	{
		map[string]string{"a": "a", "b": "b", "c": "c"},
		map[string]string{"a": "a", "c": "c"},

		[]string{},
		[]string{"b"},
		[]string{},
	},
	{
		map[string]string{"a": "a", "b": "b"},
		map[string]string{"a": "a", "b": "b", "c": "c"},

		[]string{},
		[]string{},
		[]string{"c"},
	},
	{
		map[string]string{},
		map[string]string{},

		[]string{},
		[]string{},
		[]string{},
	},
	{
		map[string]string{"a": "a"},
		map[string]string{},

		[]string{},
		[]string{"a"},
		[]string{},
	},
	{
		nil,
		map[string]string{"a": "a"},

		[]string{},
		[]string{},
		[]string{"a"},
	},
	{
		map[string]string{"a": "a"},
		nil,

		[]string{},
		[]string{"a"},
		[]string{},
	},
	{
		nil,
		nil,

		[]string{},
		[]string{},
		[]string{},
	},
}

func sliceEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, valueA := range a {
		if b[i] != valueA {
			return false
		}
	}

	return true
}

func TestMapDiff(t *testing.T) {
	for _, md := range mdTest {
		expDiff, expOnlyE, expOnlyA := util.MapDiff(md.expected, md.actual)

		if !sliceEquals(expDiff, md.diff) {
			t.Errorf("expected diff %v, got %v", md.diff, expDiff)
		}

		if !sliceEquals(expOnlyE, md.onlyExpected) {
			t.Errorf("expected onlyExpected %v, got %v", md.onlyExpected, expOnlyE)
		}

		if !sliceEquals(expOnlyA, md.onlyActual) {
			t.Errorf("expected onlyActual %v, got %v", md.onlyActual, expOnlyA)
		}
	}
}

func ExampleMapDiff() {
	expected := map[string]string{"a": "a", "b": "b"}
	actual := map[string]string{"a": "aaaa", "c": "c"}

	diff, onlyExpected, onlyActual := util.MapDiff(expected, actual)
	fmt.Println(diff)         // ["a"]
	fmt.Println(onlyExpected) // ["b"]
	fmt.Println(onlyActual)   // ["c"]

	diff, onlyExpected, onlyActual = util.MapDiff(expected, expected)
	fmt.Println(diff)         // []
	fmt.Println(onlyExpected) // []
	fmt.Println(onlyActual)   // []
}
