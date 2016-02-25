package client

import (
	"fmt"

	"github.com/alkasir/alkasir/pkg/i18n"
	"github.com/alkasir/alkasir/pkg/res"
	"github.com/thomasf/lg"
)

func loadTranslations(languages ...string) {

	for _, lang := range languages {
		filename := fmt.Sprintf("messages/%s/messages.json", lang)
		lg.V(6).Infoln("Loading translation", lang, filename)
		data, err := res.Asset(filename)
		if err != nil {
			panic(err)
		}
		err = i18n.AddBundle(lang, data)
		if err != nil {
			panic(err)
		}
	}
}
