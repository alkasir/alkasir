package osutil

import (
	"os"
	"path/filepath"
	"runtime"
)

func HomePath() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		if home == "" {
			// Store settings in same directory as the binary as a last resort,
			// this should not happen but who knows.
			home = filepath.Dir(os.Args[0])
		}
		return home
	}
	return os.Getenv("HOME")
}
