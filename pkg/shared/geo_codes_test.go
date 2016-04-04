package shared

import "testing"

func TestGeoCodes(t *testing.T) {
	legacyCCs := map[string]bool{
		"AN": true,
		"CS": true,
		"TP": true,
	}

	for _, cc := range CountryCodes {
		if _, ok := Continents[cc]; !ok {
			if _, ok := legacyCCs[cc]; !ok {
				t.Errorf("country code %s not in contientds", cc)
			}
		}
	}
}
