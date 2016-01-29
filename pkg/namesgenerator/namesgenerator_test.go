package namesgenerator

import (
	"fmt"
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
		if k != GetPronouncableName(v) {
			// fmt.Printf("GOT: %s EXPECTED: %s FROM: %s", GetPronouncableName(v), k, v)
			fmt.Printf("GOT: %s EXPECTED: %s FROM: %s \n", GetPronouncableName(v), k, v)
			t.Fail()
		}
	}

}
