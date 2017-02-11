// +build windows darwin

package browsercode

import "github.com/atotto/clipboard"

func (b *BrowserCode) CopyToClipboard() error {
	encoded, err := b.Encode()
	if err != nil {
		return err
	}
	err = clipboard.WriteAll(encoded)
	if err != nil {
		return err
	}
	return nil

}
