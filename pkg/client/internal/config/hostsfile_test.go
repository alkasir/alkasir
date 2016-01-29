package clientconfig_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	. "github.com/alkasir/alkasir/pkg/client/internal/config"
)

var testSettings = `
[local]
ControllerBindAddr = "127.0.0.1:8899"
HttpProxyBindAddr =  "localhost:9124"
SocksProxyBindAddr = "localhost:9125"
Language = "en"
CountryCode = "SE"
[[connection]]
transport = "socks5"

[transport.socks5]
bundled = true
`

func tempdir(t *testing.T) (tempdir string) {
	tempdir, err := ioutil.TempDir("", "atest")
	if err != nil {
		t.Fail()
	}
	return
}

func TestLoadNonExistentHostsfile(t *testing.T) {
	basedir := tempdir(t)
	defer os.RemoveAll(basedir)
	hf := &HostsFile{
		Name:        "testfile",
		CountryCode: "SE",
	}
	err := hf.Read(basedir)
	if err == nil {
		t.Fail()
	}
}

func TestAddHosts(t *testing.T) {
	hf := &HostsFile{
		Name:        "testfile",
		CountryCode: "SE",
	}
	hf.Add("google.com")
	if hf.Hosts[0] != "google.com" {
		t.Fail()
	}
}

func TestSaveHosts(t *testing.T) {
	basedir := tempdir(t)
	defer os.RemoveAll(basedir)
	hf := &HostsFile{
		Name:        "testfile",
		CountryCode: "SE",
	}
	hf.Add("google.com")

	err := hf.Write(basedir)
	if err != nil {
		t.Error(err)
	}
}

func TestReadHostsfile(t *testing.T) {
	basedir := tempdir(t)
	defer os.RemoveAll(basedir)

	fulldir := path.Join(basedir, "hostlists/SE")
	os.MkdirAll(fulldir, 0775)

	err := ioutil.WriteFile(
		path.Join(fulldir, "asdf.txt"),
		[]byte("alkasir.com\nsome.domain"),
		0775)

	if err != nil {
		t.Error(err)
	}
	hf := &HostsFile{
		Name:        "asdf",
		CountryCode: "SE",
	}
	err = hf.Read(basedir)

	if err != nil {
		t.Error(err)
	}
	if hf.Hosts[0] != "alkasir.com" {
		t.Fail()
	}
	if hf.Hosts[1] != "some.domain" {
		t.Fail()
	}
}
