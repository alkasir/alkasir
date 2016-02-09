// Package util stores useful helper code.
package util

import (
	"fmt"
	"regexp"
	"strings"
)

// ByteSize represents an amount of bytes. float64 is needed since division
// is required. Implements the fmt.Stringer interface.
type ByteSize float64

// Some pre-constructed file size units.
const (
	Byte ByteSize = 1 << (10 * iota)
	Kilobyte
	Megabyte
	Gigabyte
)

// String implements the fmt.Stringer interface.
func (size ByteSize) String() string {
	switch {
	case size <= Kilobyte:
		return fmt.Sprintf("%d B", int(size))
	case size <= Megabyte:
		return fmt.Sprintf("%.2f KB", size/Kilobyte)
	case size <= Gigabyte:
		return fmt.Sprintf("%.2f MB", size/Megabyte)
	default:
		return fmt.Sprintf("%.2f GB", size/Gigabyte)
	}
}

var urlRe = regexp.MustCompile(`^(?P<scheme>[^:]+)://(?P<rest>.+)`)
var slashesRe = regexp.MustCompile(`//+`)

// CleanSlashes removes extraneous slashes (like nexus.com///something), which
// Nexus' API doesn't recognize as valid. Returns an util.MalformedURLError if
// the given URL can't be parsed.
func CleanSlashes(url string) (string, error) {
	matches := urlRe.FindStringSubmatch(url)
	if matches == nil {
		return "", &MalformedURLError{url}
	}

	// scheme = matches[1] and rest = matches[2]. Clean the extraneous slashes
	return matches[1] + "://" + slashesRe.ReplaceAllString(matches[2], "/"), nil
}

// BuildFullURL builds a complete URL string in the format host/path?query,
// where query's keys and values will be formatted as k=v. Returns an
// util.MalformedURLError if the given URL can't be parsed. This function is a
// (very simplified version of url.URL.String().
func BuildFullURL(host string, path string, query map[string]string) (string, error) {
	params := []string{}

	for k, v := range query {
		params = append(params, k+"="+v)
	}

	if len(params) == 0 {
		return CleanSlashes(host + "/" + path)
	}

	return CleanSlashes(host + "/" + path + "?" + strings.Join(params, "&"))
}

// MalformedURLError is returned when the given URL could not be parsed.
type MalformedURLError struct {
	URL string // e.g. http:/:malformed.url.com
}

// Error implements the error interface.
func (err MalformedURLError) Error() string {
	return fmt.Sprintf("Malformed URL: %v", err.URL)
}

//MapDiff calculates the difference between two maps. It returns three values:
//
//* diff: a slice of strings, holding the keys in both with different values;
//
//* onlyExpected: a slice of keys only in expected;
//
//* onlyActual: a slice of keys only in actual.
//
// When the maps are equal, all three slices will be empty.
func MapDiff(expected map[string]string, actual map[string]string) (diff []string, onlyExpected []string, onlyActual []string) {
	keysSeen := map[string]bool{}

	for keyExp, valueExp := range expected {
		keysSeen[keyExp] = true // marking keyExp to avoid redoing work

		valueAct, ok := actual[keyExp]
		if !ok { // keyExp isn't in actual
			onlyExpected = append(onlyExpected, keyExp)
		} else if valueAct != valueExp { // expected and actual differ
			diff = append(diff, keyExp)
		} // else the keys and values match, nothing to do
	}

	for keyAct, valueAct := range actual {
		if keysSeen[keyAct] { // already processed, move along
			continue
		}

		valueExp, ok := expected[keyAct]
		if !ok { // keyAct isn't in actual
			onlyActual = append(onlyActual, keyAct)
		} else if valueExp != valueAct { // expected and actual differ
			diff = append(diff, keyAct)
		}
	}

	return // diff, onlyExpected, onlyActual
}
