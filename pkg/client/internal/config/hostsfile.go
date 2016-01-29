// This is currently the placs for the registry of blocked urls
package clientconfig

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/thomasf/lg"
)

// HostsFile represents a text file with one url per line.
type HostsFile struct {
	Name        string
	Hosts       []string
	CountryCode string
}

// Add an entry to a HostsFile
func (h *HostsFile) Add(host string) {
	// todo.. this is a good place to check the format of the string
	host = strings.TrimSpace(host)
	for _, h := range h.Hosts {
		if h == host {
			return
		}
	}
	h.Hosts = append(h.Hosts, host)
}

// Remove an entry from an HostsFile
func (h *HostsFile) Remove(host string) {
	var hosts []string
	host = strings.TrimSpace(host)
	for _, h := range h.Hosts {
		if h != host {
			hosts = append(hosts, h)
		}
	}
	h.Hosts = hosts
}

func (h *HostsFile) fullpath(basedir string) string {
	return path.Join(basedir, "hostlists", h.CountryCode, h.Name+".txt")
}

// Read the HostList from file
func (h *HostsFile) Read(basedir string) (err error) {
	filename := h.fullpath(basedir)
	lg.V(5).Info("Reading hosts file", filename)
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	err = scanner.Err()
	if err != nil {
		return err
	}
	h.Hosts = lines
	return
}

// Write the HostList to file
func (h *HostsFile) Write(basedir string) (err error) {
	filename := h.fullpath(basedir)
	dirname := path.Dir(filename)
	os.MkdirAll(dirname, 0755)
	if err != nil {
		return err
	}

	// TODO: in place rewrite of files are maybe not a good idea..
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, line := range h.Hosts {
		fmt.Fprintln(w, line)
	}
	return w.Flush()
}
