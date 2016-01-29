package client

var (
	// LanguageOptions = []string{"en", "ar", "zh", "sv", "fa"}
	LanguageOptions = []string{"en", "ar", "fa"}
	validLanguages  map[string]bool
)

func ValidLanguage(lang string) bool {
	_, ok := validLanguages[lang]
	return ok
}

func init() {
	validLanguages = make(map[string]bool, 0)
	for _, v := range LanguageOptions {
		validLanguages[v] = true
	}
}
