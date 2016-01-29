// GUI package, requires cgo/platform specific stuff.
//
//
package wm

import (
	"runtime"
	"sync"

	"github.com/atotto/clipboard"
	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/client/ui"
	"github.com/alkasir/alkasir/pkg/client/ui/systray"
	"github.com/alkasir/alkasir/pkg/i18n"
	"github.com/alkasir/alkasir/pkg/res"
)

type WMGUI struct {
	onReady                func()
	quit, copy, open, help *systray.MenuItem
	langMu                 sync.Mutex
	langSet                bool
}

func New() *WMGUI {
	return &WMGUI{}
}

func (w *WMGUI) init() {
	icn := "img/alkasir-icon-16x16.png"
	if runtime.GOOS == "windows" {
		icn = "img/alkasir-icon-128x128.ico"
	}

	icon, err := res.Asset(icn)
	if err != nil {
		panic(err)
	}
	systray.SetIcon(icon)
	w.onReady()
}

func (w *WMGUI) Run(onReady func()) error {
	w.onReady = onReady
	systray.Run(w.init)
	return nil
}

func (w *WMGUI) Done() error {
	systray.Quit()
	return nil
}

func (w *WMGUI) WriteClipboard(msg string) error {
	return clipboard.WriteAll(msg)
}

func (w *WMGUI) Language(lang string) error {
	w.langMu.Lock()
	defer w.langMu.Unlock()
	T, err := i18n.Tfunc(lang)
	if err != nil {
		return err
	}
	quitMsg := T("quit_alkasir")
	helpMsg := T("action_help")
	copyMsg := T("copy_browser_code_to_clipboard")
	openMsg := T("open_in_browser")
	if !w.langSet {
		w.open = systray.AddMenuItem(openMsg, openMsg)
		w.help = systray.AddMenuItem(helpMsg, helpMsg)
		w.copy = systray.AddMenuItem(copyMsg, copyMsg)
		w.quit = systray.AddMenuItem(quitMsg, quitMsg)
		w.langSet = true

		go func() {
			for {
				select {
				case <-w.quit.ClickedCh:
					lg.Infoln("quit clicked")
					go func() {
						ui.Actions.Quit <- true
					}()
				case <-w.copy.ClickedCh:
					go func() {
						lg.Infoln("copy to clipboard clicked")
						ui.Actions.CopyBrowserCodeToClipboard <- true
					}()
				case <-w.help.ClickedCh:
					go func() {
						lg.Infoln("help clicked")
						ui.Actions.Help <- true
					}()
				case <-w.open.ClickedCh:
					go func() {
						lg.Infoln("open browser clicked")
						ui.Actions.OpenInBrowser <- true
					}()
				}
			}
		}()
		return nil
	}
	w.quit.SetTitle(quitMsg)
	w.quit.SetTooltip(quitMsg)
	w.help.SetTitle(helpMsg)
	w.help.SetTooltip(helpMsg)
	w.copy.SetTitle(copyMsg)
	w.copy.SetTooltip(copyMsg)
	w.open.SetTitle(openMsg)
	w.open.SetTooltip(openMsg)
	return nil
}
