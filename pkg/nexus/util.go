package nexus

import (
	"io"
	"net/http"
	"os"
)

func download(dst, url string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

func isExecutable(d os.FileInfo) bool {
	if m := d.Mode(); !m.IsDir() && m&0111 != 0 {
		return true
	}
	return false
}
