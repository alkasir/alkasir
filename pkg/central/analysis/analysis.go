package analysis

import (
	"encoding/json"
	"time"

	"github.com/alkasir/alkasir/pkg/central/db"
	"github.com/alkasir/alkasir/pkg/measure"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/thomasf/lg"
)

var (
	sessionFetchC   = make(chan shared.SuggestionToken, 0)
	sampleAnalysisC = make(chan []db.Sample, 0)
	hostPublishC    = make(chan db.Sample, 0)
	hostDenyC       = make(chan db.Sample, 0)
)

func sessionFetcher(clients db.Clients) {
loop:
	for token := range sessionFetchC {
		if token == "" {
			lg.Errorln("empty token received, skipping")
			continue loop

		}
		samples, err := clients.DB.GetSessionSamples(token)
		if err != nil {
			lg.Error(err)
			continue loop
		}
		sampleAnalysisC <- samples
	}
}

func samplesAnalyzer() {
loop:
	for samples := range sampleAnalysisC {
		var (
			newTokenSample *db.Sample
			clientSamples  = make(map[string]db.Sample, 0)
			centralSamples = make(map[string]db.Sample, 0)
		)

		// organize input data
		for _, s := range samples {
			switch s.Type {
			case "NewClientToken":
				if newTokenSample != nil {
					lg.Error("got more than one newTokenSample, aborting")
					continue loop
				}
				newTokenSample = &s
			case "HTTPHeader", "DNSQuery":
				switch s.Origin {
				case "Central":
					centralSamples[s.Type] = s
				case "Client":
					clientSamples[s.Type] = s
				}
			default:
				lg.Errorf("dont know how to handle %d %s, skipping", s.ID, s.Type)
			}
		}

		// validate that wanted data types are available
		if newTokenSample == nil {
			lg.Errorln("No newTokenSample, aborting")
			continue loop
		}
		if !shared.AcceptedHost(newTokenSample.Host) {
			lg.Warningln("not accepted host id:", newTokenSample.ID, newTokenSample.Host)
			continue loop
		}
		for _, smap := range []map[string]db.Sample{clientSamples, centralSamples} {
			for _, stype := range []string{"HTTPHeader"} {
				if _, ok := smap[stype]; !ok {
					lg.Errorf("missing %s, cannot analyse", stype)
					continue loop
				}
			}
		}

		// parse data
		clientSample := clientSamples["HTTPHeader"]
		var clientHeader measure.HTTPHeaderResult
		centralSample := centralSamples["HTTPHeader"]
		var centralHeader measure.HTTPHeaderResult

		if err := json.Unmarshal(clientSample.Data, &clientHeader); err != nil {
			lg.Errorln(err)
			continue loop
		}
		if err := json.Unmarshal(centralSample.Data, &centralHeader); err != nil {
			lg.Error(err)
			continue loop
		}

		score := scoreHTTPHeaders(clientHeader, centralHeader)
		lg.V(10).Infof("session:%s %s", newTokenSample.Token, score)
		if score.Score() >= 0.5 {
			lg.Infoln("publishing session", newTokenSample.Token)
			hostPublishC <- *newTokenSample
		} else {
			lg.Infoln("not publishing session", newTokenSample.Token)
		}
	}
}

func hostPublisher(clients db.Clients) {
	for sample := range hostPublishC {
		err := clients.DB.PublishHost(sample)
		if err != nil {
			lg.Error(err)
		}
	}
}

func StartAnalysis(clients db.Clients) {

	tick := time.NewTicker(5 * time.Second)
	lastID, err := clients.DB.GetLastProcessedSampleID()
	if err != nil {
		lg.Warningln(err)
	}

	for n := 0; n < 4; n++ {
		go sessionFetcher(clients)
		go samplesAnalyzer()
		go hostPublisher(clients)
	}

	lg.Infof("starting analysis from sample ID %d", lastID)

	lastPersistedID := lastID

	for range tick.C {
		results, err := clients.DB.GetSamples(uint64(lastID), "")
		if err != nil {
			lg.Errorf("database err (skipping): %v", err)
			continue
		}
		n := 0
		start := time.Now()
		for s := range results {
			n++
			if s.ID > lastID {
				lastID = s.ID
			}

			if s.Origin == "Central" && s.Type == "HTTPHeader" {
				sessionFetchC <- s.Token
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

}
