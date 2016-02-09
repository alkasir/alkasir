package cliui

import (
	"log"
	"sync"
)

type CLIUI struct{}

func lg(cmd, msg string) {
	log.Printf("*** CLIUI [%s] %s ", cmd, msg)
}

func New() *CLIUI {
	return &CLIUI{}
}

var wg sync.WaitGroup

func (c *CLIUI) Run(onReady func()) error {
	wg.Add(1)
	onReady()
	wg.Wait()
	return nil
}

func (c *CLIUI) SetURL(url string) {
	lg("seturl", url)
}

func (c *CLIUI) Done() error {
	wg.Done()
	return nil
}

func (c *CLIUI) Language(lang string) error {
	return nil
}

func (c *CLIUI) WriteClipboard(msg string) error {
	lg("WriteClipboard", "clipboard not supported in cliui, ignoring.")
	return nil
}
