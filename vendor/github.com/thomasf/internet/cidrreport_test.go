package internet

import (
	"os"
	"testing"
)

func TestParseCidrreport(t *testing.T) {
	file, err := os.Open("testdata/autnums.sample.txt")
	n := 0
	err = parseReport(file, func(*ASDescription) error {
		n++
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	if n != 139 {
		t.Errorf("Expected 139 rows, got %d", n)
	}
}

func TestParseBrokenCidrreport(t *testing.T) {
	file, err := os.Open("testdata/autnums.invalid.sample.txt")
	err = parseReport(file, func(*ASDescription) error { return nil })
	if _, ok := err.(ParseError); !ok {
		t.Fatalf("expected parse error, got %s", err)
	}
}

func TestParseBrokenLineCidrreport(t *testing.T) {
	file, err := os.Open("testdata/autnums.invalid.sample2.txt")
	err = parseReport(file, func(*ASDescription) error { return nil })

	if _, ok := err.(ParseError); !ok {
		t.Fatalf("expected parse error, got %s", err)
	}
}
