// +build linux freebsd netbsd openbsd solaris

package browsercode

import (
	"errors"
	"sync"
	"time"

	"github.com/thomasf/lg"
	"github.com/tvdburgt/passman/clipboard"
)

var (
	initOnce sync.Once
	initErr  error
)

func (b *BrowserCode) CopyToClipboard() error {

	setup := func() {
		initErr = clipboard.Setup()
	}
	initOnce.Do(setup)
	if initErr != nil {
		lg.Errorln(initErr)
		return initErr
	}

	encoded, err := b.Encode()
	if err != nil {
		return err
	}

	sC, eC := clipboard.Put([]byte(encoded))
	select {
	case <-sC:
		return nil
	case err := <-eC:
		return err
	case <-time.After(2 * time.Second):
		return errors.New("time out setting clip board")
	}

	return nil

}
