package client

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/alkasir/alkasir/pkg/osutil"
	"github.com/alkasir/alkasir/pkg/res/chrome"
	"github.com/thomasf/lg"
)

var saveChromeMu sync.Mutex

func saveChromeExtension() error {
	saveChromeMu.Lock()
	defer saveChromeMu.Unlock()
	exportPath := filepath.Join(osutil.HomePath(), "AlkasirChromeExtension")
	if err := os.RemoveAll(exportPath); err != nil {
		lg.Errorln(err)
	}
	err := chrome.RestoreAssets(exportPath, "")
	if err != nil {
		return err
	}
	return nil

}
