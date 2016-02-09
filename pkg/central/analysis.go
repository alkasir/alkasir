package central

import (
	"time"

	"github.com/thomasf/lg"
	"github.com/alkasir/alkasir/pkg/central/db"
	"github.com/alkasir/alkasir/pkg/shared"
)

func startAnalysis(clients db.Clients) {
	go func() {
		tick := time.NewTicker(10 * time.Second)

		lastID, err := clients.DB.GetLastProcessedSampleID()
		if err != nil {
			lg.Warningln(err)
		}
		lg.Infof("starting analysis from sample ID %d", lastID)

		lastPersistedID := lastID

		for range tick.C {
			results, err := clients.DB.GetSamples(uint64(lastID), "")
			if err != nil {
				lg.Fatal(err)
			}
			n := 0
			start := time.Now()
		loop:
			for s := range results {
				n++
				if s.ID > lastID {
					lastID = s.ID
				}
				if s.Type == "NewClientToken" {
					if !shared.AcceptedHost(s.Host) {
						lg.Warningln("not accepted host id:", s.ID, s.Host)
						continue loop
					}
					err := clients.DB.PublishHost(s)
					if err != nil {
						lg.Warning(err)
					}
				}
			}
			if n != 0 && lg.V(15) {
				lg.Infof("processed %d samples in %s", n, time.Since(start).String())
			}
			if lastID != lastPersistedID {
				err = clients.DB.SetLastProcessedSampleID(lastID)
				if err != nil {
					lg.Errorln(err)
				} else {
					lastPersistedID = lastID
				}
			}
		}
	}()
}
