package client

import (
	"os"
	"path/filepath"

	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/osutil"
	"github.com/alkasir/alkasir/pkg/res/chrome"
)

func saveChromeExtension() error {
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
