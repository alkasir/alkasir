package res

import (
	"path"
	"runtime"
)

// String for go-bindata dev mode
var rootDir string

func init() {
	_, filename, _, _ := runtime.Caller(1)
	rootDir = path.Join(path.Dir(filename), "../../res")
}
