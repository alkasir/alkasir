package displayname

import (
	"fmt"
	"hash/crc32"
	"testing"
)

func TestNameGenerator(t *testing.T) {
	for k, v := range map[string]string{
		"ink leaf freckle ivory":       "J0Ijoib2JmczQiLCJzIjoiXCJjZXJ0PUdzVFAxVmNwcjBJeE9aUkNnUHZ0Z1JsZVJWQzFTRmpYWVhxSDVvSEhYVFJ4M25QVnZXN2xHK2RkREhKWmw4YjBOVFZ1VGc7aWF0LW1vZGU9MFwiIiwiYSI6IjEzOS4xNjIuMjIxLjEyMzo0NDMifQXGp1dEeFrBpmi1SfSmdGQdz0XVeW2v6hAtL4bSEYuqtHcjJFw3XyblvRfQ7nAm6bATWDyoxBLlhHWo8jy6LjHNU5O5ZebO8YfySoijD8S21zVxhO7UtR6Spy3RNuzuxa==",
		"fir holly crystal cerulean":   "pVIDlNXalDcqN5BaPVJqNJTHQAbaiyXja9pUARhLd1VAQ6xKGmI4nacifGNpjSol=",
		"granite dandy shag red":       "G1kK7N3w",
		"spring cherry clear brown":    "1en3TYtt",
		"vivid narrow well cyan":       "1en3TYts",
		"plume winter bow iridescent":  "a",
		"shadow tree cat":              "aa",
		"tree snow alabaster viridian": "b",
		"sand black cactus ivory":      "1en3TYt",
	} {
		PN := FromString(v)
		if k != PN {
			fmt.Printf("GOT: %s EXPECTED: %s FROM: %s \n", PN, k, v)
			t.Fail()
		}
	}
}

const (
	Some = 1e4
)

func BenchmarkDefaultGenerator(b *testing.B) {
	str := string(make([]byte, Some))
	for i := 0; i < b.N; i++ {
		_ = FromString(str)
	}
}

// just to have a reference for the time spent in the hash function
func BenchmarkCRC32(b *testing.B) {
	str := string(make([]byte, Some))
	c := crc32.NewIEEE()
	for i := 0; i < b.N; i++ {
		c.Reset()
		_, err := c.Write([]byte(str))
		if err != nil {
			panic(err)
		}
		c.Sum(nil)
	}
}

func BenchmarkGenerator(b *testing.B) {
	str := string(make([]byte, Some))
	g := Generator{
		Hash:  crc32.NewIEEE(),
		Words: defaultWords,
	}
	for i := 0; i < b.N; i++ {
		_, err := g.FromString(str)
		if err != nil {
			panic(err)
		}
	}
}
