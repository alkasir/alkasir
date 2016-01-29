// api for ui module(s)
package ui

import (
	"log"
	"sync"

	"github.com/thomasf/lg"
)

var Actions = struct {
	Quit, CopyBrowserCodeToClipboard, OpenInBrowser, Help chan interface{}
}{
	CopyBrowserCodeToClipboard: make(chan interface{}, 0),
	Quit:          make(chan interface{}, 0),
	OpenInBrowser: make(chan interface{}, 0),
	Help:          make(chan interface{}, 0),
}

// ui us te definition of a UI service
type ui interface {
	Run(onReady func()) error
	Done() error
	WriteClipboard(msg string) error

	Language(lang string) error
}

// u is the reference of the UI implementation used

var (
	u     ui
	setui sync.Once
)

// Set sets the ui implementation
func Set(ui ui) {
	setui.Do(func() {
		u = ui
	})
}

// Init wrapper
func Run(onReady func()) error {
	lg.V(19).Infoln("entering ui.Run")
	err := u.Run(onReady)
	lg.V(19).Infoln("leaving ui.Run")
	return err
}

func WriteClipboard(msg string) error {
	return u.WriteClipboard(msg)
}

func Language(lang string) error {
	return u.Language(lang)
}

func Done() {
	lg.V(19).Infoln("entering ui.Done")
	u.Done()
	lg.V(19).Infoln("leaving ui.Done")
}

func Notify(msg string) {
	log.Println("ui.Notify is a noop righjt now, the call should probably be removed:", msg)

}
