package internet

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/garyburd/redigo/redis"
	"github.com/osrg/gobgp/packet/bgp"
	"github.com/osrg/gobgp/packet/mrt"

)

// RefreshBGPDump ensures that the latest dump available is the one which is installed.
func RefreshBGPDump(conn redis.Conn) (int, error) {
	for _, b := range []BGPDump{
		{Date: time.Now()},
		{Date: time.Now().Add(-time.Duration(time.Hour * 24))},
	} {
		err := b.Download()
		if err != nil {
			return 0, err
		}
		if b.IsDownloaded() {
			return b.Import(conn)
		}
	}
	return 0, nil
}

// BGPDump encapuslates downloading and importing of BGP dumps.
type BGPDump struct {
	Date time.Time
}

// Import stores the contents of a downloaded BGP dump into a redis server.
// -1 is returned if the dump is alredy imported into redis.
func (b *BGPDump) Import(conn redis.Conn) (int, error) {
	alreadyImported, err := redis.Bool(conn.Do("SISMEMBER", "i2a:imported_dates", b.day()))
	if err != nil {
		return 0, err
	}
	if alreadyImported {
		return -1, nil
	}
	n, err := b.parseBGPDump(conn)
	return n, err

}

// IsDownloaded returns true if the BGPDump archive is downloaded locally.
func (b *BGPDump) IsDownloaded() bool {
	p := b.Path()
	if _, err := os.Stat(p); err == nil {
		return true
	}
	return false
}

// Download fetches an bgpdump archive from http://data.ris.ripe.net/rrc00.
// A http 404 status code does not generate an error, the isDownloaded() to check success after fetching.
// Download returns early with no error if the file already is downloaded to disk.
func (b *BGPDump) Download() error {
	dt := b.Date
	dumpDir := b.dir()
	err := os.MkdirAll(dumpDir, 0777)
	if err != nil {
		return err
	}
	if b.IsDownloaded() {
		return nil
	}
	err = os.MkdirAll(filepath.Join(dataDir, "spool"), 0777)
	if err != nil {
		return err
	}
	tempFile, err := ioutil.TempFile(
		filepath.Join(dataDir, "spool"), b.day())
	if err != nil {
		return err
	}
	defer tempFile.Close()
	dlURL := fmt.Sprintf(
		"http://data.ris.ripe.net/rrc00/%s/bview.%s.%s.gz",
		dt.Format("2006.01"), b.day(), "0000")

	resp, err := http.Get(dlURL)
	if err != nil {
		return err
	}

	// Dumps from ??? to 2010-06-14 are named timestamped 2359 so do a check
	// for that if 0000 fails. For very early dumps the format is not static so those will fail.
	if resp.StatusCode == 404 && dt.Before(time.Date(2010, 06, 15, 0, 0, 0, 0, time.UTC)) {
		// log.Printf("trying different url, got 404 for %s", dlURL)
		dlURL = fmt.Sprintf(
			"http://data.ris.ripe.net/rrc00/%s/bview.%s.%s.gz",
			dt.Format("2006.01"), b.day(), "2359")
		resp, err = http.Get(dlURL)
		if err != nil {
			return err
		}
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 404 {
			// log.Printf("Skipping download, got 404 for %s", dlURL)
			return nil
		}
		return fmt.Errorf("Got http status code %s response for %s", resp.Status, dlURL)
	}
	// log.Printf("Downloading %s\n", dlURL)

	defer resp.Body.Close()
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return err

	}
	err = os.Rename(tempFile.Name(), b.Path())
	if err != nil {
		return err

	}
	return nil

}

func (b *BGPDump) parseBGPDump(conn redis.Conn) (int, error) {
	day := b.day()
	n := 0

	f, err := os.Open(b.Path())
	if err != nil {
		return 0, err
	}

	gzipReader, err := gzip.NewReader(f)
	if err != nil {
		return n, fmt.Errorf("couldn't create gzip reader: %v", err)
	}
	scanner := bufio.NewScanner(gzipReader)
	scanner.Split(mrt.SplitMrt)
	count := 0

	indexTableCount := 0
entries:
	for scanner.Scan() {
		count++
		data := scanner.Bytes()

		hdr := &mrt.MRTHeader{}
		errh := hdr.DecodeFromBytes(data[:mrt.MRT_COMMON_HEADER_LEN])
		if err != nil {
			return 0, errh
		}

		msg, err := mrt.ParseMRTBody(hdr, data[mrt.MRT_COMMON_HEADER_LEN:])
		if err != nil {
			log.Printf("could not parse mrt body: %v", err)
			continue entries
		}

		if msg.Header.Type != mrt.TABLE_DUMPv2 {
			return 0, fmt.Errorf("unexpected message type: %d", msg.Header.Type)
		}

		switch mtrBody := msg.Body.(type) {
		case *mrt.PeerIndexTable:
			indexTableCount++
			if indexTableCount != 1 {
				return 0, fmt.Errorf("got >1 PeerIndexTable")
			}

		case *mrt.Rib:
			prefix := mtrBody.Prefix
			if len(mtrBody.Entries) < 0 {
				return 0, fmt.Errorf("no entries")
			}

			for _, entry := range mtrBody.Entries {
			attrs:
				for _, attr := range entry.PathAttributes {
					switch attr := attr.(type) {
					case *bgp.PathAttributeAsPath:
						if len(attr.Value) < 1 {
							continue attrs
						}
						if v, ok := attr.Value[0].(*bgp.As4PathParam); ok {
							if len(v.AS) < 0 {
								continue attrs
							}
							conn.Send("HSET", fmt.Sprintf("i2a:%s", prefix), day, v.AS[len(v.AS)-1])
							n++
							if n%10000 == 0 {
								err := conn.Flush()
								if err != nil {
									return 0, err
								}
							}
							continue entries
						}
					}
				}
			}
		default:
			return 0, fmt.Errorf("unsupported message %v %s", mtrBody, spew.Sdump(msg))
		}
	}
	conn.Send("SADD", "i2a:imported_dates", day)
	err = conn.Flush()
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (b *BGPDump) parseBGPCSV(r io.Reader, conn redis.Conn) (int, error) {
	day := b.day()
	s := bufio.NewScanner(r)
	n := 0
	var asn string
	for s.Scan() {
		cols := strings.Split(s.Text(), "|")
		if len(cols) < 7 {
			return n, ParseError{
				Message: "too few columns",
				Path:    filepath.Base(b.Path()),
				LineNum: n,
				Line:    s.Text(),
			}
		}
		block := cols[5]

		if _, ok := asn12654blocks[block]; ok {
			asn = "12654"
		} else {
			asPath := cols[6]
			asns := strings.Split(asPath, " ")
			asn = asns[len(asns)-1]
			if asn == "" {
				return n, ParseError{
					Message: "no ASPATH data",
					Path:    filepath.Base(b.Path()),
					LineNum: n,
					Line:    s.Text(),
				}
			}
		}
		conn.Send("HSET", fmt.Sprintf("i2a:%s", block), day, asn)
		n++
		if n%10000 == 0 {
			err := conn.Flush()
			if err != nil {
				return 0, err
			}
		}
	}
	conn.Send("SADD", "i2a:imported_dates", day)
	err := conn.Flush()
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Path returns the absolute path to the target archive dump download file.
func (b *BGPDump) Path() string {
	return filepath.Join(
		b.dir(), fmt.Sprintf("%s.gz", b.Date.Format("20060102")))
}

func (b *BGPDump) dir() string {
	return filepath.Join(
		dataDir, "cache", b.Date.Format("200601"))
}

func (b *BGPDump) day() string {
	return b.Date.Format("20060102")
}
