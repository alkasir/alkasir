// Copyright 2012 Rémy Oudompheng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitdelta

import (
	"bytes"
	"io/ioutil"
	"testing"

	"fmt"
	"math/rand"
	"strconv"
)

func mustRead(t *testing.T, n string) []byte {
	s, err := ioutil.ReadFile(n)
	if err != nil {
		t.Fatalf("reading %s: %s", n, err)
	}
	return s
}

func TestPatch(t *testing.T) {
	s1 := mustRead(t, "testdata/golden.old")
	s2 := mustRead(t, "testdata/golden.new")
	// Patch created using the test-delta utility from git sources.
	p := mustRead(t, "testdata/golden.delta")

	s, err := Patch(s1, p)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(s, s2) {
		t.Errorf("difference: got %q, expect %q", s, s2)
	}
}

func TestDiff(t *testing.T) {
	s1 := mustRead(t, "testdata/golden.old")
	s2 := mustRead(t, "testdata/golden.new")
	p := mustRead(t, "testdata/golden.delta")

	patch := Diff(s1, s2)
	if len(p) != len(patch) {
		// we don't expect to get the exact same delta.
		t.Logf("expected %d bytes, got %d bytes", len(p), len(patch))
	}

	s2test, err := Patch(s1, patch)
	if err != nil {
		t.Errorf("produced invalid patch: %s", err)
	}
	if !bytes.Equal(s2, s2test) {
		t.Errorf("patch doesn't reconsitute golden.new")
	}
}

func genData(size int) (s1, s2 []byte) {
	gcd := func(a, b uint32) uint32 {
		for {
			switch {
			case b == 0:
				return a
			case a == 0:
				return b
			case a < b:
				a, b = b, a
			default:
				a, b = b, a%b
			}
		}
	}

	s1 = make([]byte, 0, size)
	s2 = make([]byte, 0, size)

	N := size / 9
	P := uint32(0x1fffffff)
	if size < 1<<20 {
		P = uint32(0x7ff)
	}

	// s1 is N random numbers.
	rnd := rand.NewSource(42)
	for i := 0; i < N; i++ {
		s1 = strconv.AppendInt(s1, rnd.Int63()&0xffffffff, 16)
		s1 = append(s1, '\n')
	}

	// s2 is the same numbers but if gcd(n, 0x1fffffff) != 1
	// s2 is factored as gcd * (s2/gcd)
	rnd = rand.NewSource(42)
	for i := 0; i < N; i++ {
		n := uint32(rnd.Int63())
		gcd := gcd(n, P)
		if gcd == 1 {
			s2 = strconv.AppendInt(s2, int64(n), 16)
			s2 = append(s2, '\n')
		} else {
			s2 = append(s2, fmt.Sprintf("%x %x\n", gcd, n/gcd)...)
		}
	}
	return s1, s2
}

func benchmarkDiff(b *testing.B, size int) {
	s1, s2 := genData(size)
	b.ResetTimer()
	var n int
	for i := 0; i < b.N; i++ {
		p := Diff(s1, s2)
		n = len(p)
	}
	b.SetBytes(int64(len(s1) + len(s2)))
	b.Logf("input: %d, %d bytes, patch: %d bytes", len(s1), len(s2), n)
}

func benchmarkPatch(b *testing.B, size int) {
	s1, s2 := genData(size)
	p := Diff(s1, s2)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Patch(s1, p)
	}
	b.SetBytes(int64(len(s1) + len(s2)))
	b.Logf("input: %d bytes, patch: %d bytes, out: %d bytes", len(s1), len(p), len(s2))
}

func BenchmarkDiffSmall(b *testing.B)  { benchmarkDiff(b, 4<<10) }
func BenchmarkDiffMedium(b *testing.B) { benchmarkDiff(b, 64<<10) }
func BenchmarkDiffLarge(b *testing.B)  { benchmarkDiff(b, 8<<20) }

func BenchmarkPatchSmall(b *testing.B)  { benchmarkPatch(b, 4<<10) }
func BenchmarkPatchMedium(b *testing.B) { benchmarkPatch(b, 64<<10) }
func BenchmarkPatchLarge(b *testing.B)  { benchmarkPatch(b, 8<<20) }
