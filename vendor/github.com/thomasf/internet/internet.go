/*

Package internet produces queryable information about the internet by
fetching and downloading ripe BGP dumps and cidr-report.org data into redis
databases.

THE APIs WILL BE CONSIDERED STABLE WHEN THIS MESSAGE IS REMOVED IN MAY 2015.

Features

Download BGP table dumps from http://data.ris.ripe.net/rrc00. Entries are
stored into redis for current and historical IP address to AS Number lookup.

Download http://www.cidr-report.org/as2.0/autnums.html (controlled to once per
day). Entries are stored in redis for current and historical AS Number to AS
Description lookup.

All downloads are cached so that databases can be rebuilt easily.

Golang redis query clients are also also included.

Pre requirements

BGPDump: http://www.ris.ripe.net/source/bgpdump/. Download, compile and install
it somewhere into PATH.

A Redis server.


Acknowledgments

Basic design for the IP2ASN history and ASN2ASDescription parts were inspired
from https://github.com/CIRCL/IP-ASN-history and
https://github.com/CIRCL/ASN-Description-History.

*/
package internet

import (
	"fmt"
	"os"
	"path/filepath"
)

type ParseError struct {
	Message string
	Path    string
	LineNum int
	Line    string
}

func (p ParseError) Error() string {
	return fmt.Sprintf("parse error: %s in %s:%d: %s", p.Message, p.Path, p.LineNum, p.Line)
}

var dataDir = filepath.Join(os.TempDir(), "internet")

// SetDataDir sets the storage directory for downloads cache and temporary files.
func SetDataDir(path string) {
	dataDir = path
}
