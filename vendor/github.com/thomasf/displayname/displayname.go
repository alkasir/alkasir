package displayname

import (
	"hash"
	"hash/crc32"
	"math/big"
	"strings"
)

// Generator generates human readable display names from data by combining a
// hash function and a word list.
//
// The return value is the Hash sum in the base of the length of the Words
// slice with all words. This package is NOT meant to provide any kind of
// security, crc32 is used in the default implementation since it produces a 4
// word string with a small word list.
type Generator struct {
	hash.Hash
	Words []string
	b     *big.Int
}

// String returns the code name for the current hash value being held.
func (g Generator) String() string {
	if g.b == nil {
		g.b = &big.Int{}
	}
	sum := g.Sum(nil)
	g.b.SetBytes(sum)
	return toBase(g.b, g.Words)
}

// FromString returns the display name of the input string.
func (g Generator) FromString(s string) (string, error) {
	g.Reset()
	_, err := g.Write([]byte(s))
	if err != nil {
		return "", err
	}
	return g.String(), nil
}

// FromString returns a display name based on a crc32 hash and interal wordlist.
// Returned names are composed of 4 words.
func FromString(s string) string {
	g := Generator{
		Hash:  crc32.NewIEEE(),
		Words: defaultWords,
	}
	res, err := g.FromString(s)
	if err != nil {
		panic(err)
	}
	return res
}

// toBase converts a *big.Int value (in base 10) into a given destination
// base. The result of the conversion is returned as a string.
func toBase(bi *big.Int, destBase []string) string {
	// Hack in order to "clone" the big.Int and avoid changing it.
	src := big.NewInt(0)
	src.Add(bi, big.NewInt(0))

	if big.NewInt(0).Cmp(src) == 0 {
		return destBase[0]
	}

	var digits []string
	numericBase := big.NewInt(int64(len(destBase)))

	// Keep going while bi is greater than 0.
	for src.Cmp(big.NewInt(0)) > 0 {
		remainder := big.NewInt(0).Rem(src, numericBase)
		src.Div(src, numericBase)
		digits = append(digits, destBase[remainder.Int64()])
	}

	return strings.Join(digits, " ")
}
