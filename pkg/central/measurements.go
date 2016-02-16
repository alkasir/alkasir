package central

import (
	"errors"
	"fmt"
	"time"

	"github.com/alkasir/alkasir/pkg/central/db"
	"github.com/alkasir/alkasir/pkg/measure"
	"github.com/alkasir/alkasir/pkg/measure/sampleorigins"
	"github.com/alkasir/alkasir/pkg/measure/sampletypes"
	"github.com/alkasir/alkasir/pkg/shared"
	"github.com/thomasf/lg"
)

type centralMeasurer struct {
	token     shared.SuggestionToken
	measurers []measure.Measurer
}

var (
	requestMeasurements = make(chan centralMeasurer, 5000)
)

// PreparedSample .
type PreparedSample struct {
	lastUpdated time.Time
	s           db.Sample
}

func (p *PreparedSample) Update(dbclients db.Clients) error {
	IP := shared.GetPublicIPAddr()
	if IP == nil {
		return errors.New("could not get own public ip address")
	}

	// resolve ip to asn.
	var ASN int
	ASNres, err := dbclients.Internet.IP2ASN(IP)
	if err != nil {
		lg.Errorln(err.Error())
		return err
	}
	if ASNres != nil {
		ASN = ASNres.ASN
	} else {
		lg.Warningf("no ASN lookup result for IP: %s ", IP)
		return fmt.Errorf("no ASN lookup result for IP: %s ", IP)
	}

	// resolve ip to country code.
	countryCode := dbclients.Maxmind.IP2CountryCode(IP)

	s := db.Sample{
		CountryCode: countryCode,
		ASN:         ASN,
	}

	p.lastUpdated = time.Now()
	p.s = s
	return nil
}

func startMeasurer(dbclients db.Clients) {

	for n := 0; n < 10; n++ {

		go func() {
			var ps PreparedSample
			err := ps.Update(dbclients)
			if err != nil {
				lg.Error("could not resolve public ip address", err)
			}

			lg.V(5).Infoln("starting measurer")
			for r := range requestMeasurements {
				lg.V(50).Infoln("got measurement", r)
				if ps.lastUpdated.Before(time.Now().Add(-time.Hour * 5)) {
					lg.V(15).Info("updating prepared sample", ps)
					err := ps.Update(dbclients)
					if err != nil {
						lg.Warning(err)
					}

				}
			measurerLoop:
				for _, v := range r.measurers {
					measurement, err := v.Measure()
					if err != nil {
						lg.Errorf("could not measure:%v error:%s", v, err.Error())
						continue measurerLoop
					}
					switch measurement.Type() {
					case sampletypes.DNSQuery, sampletypes.HTTPHeader:

						data, err := measurement.Marshal()
						if err != nil {
							lg.Errorf("could not decode %v error:%s", measurement, err.Error())
							continue measurerLoop
						}
						err = dbclients.DB.InsertSample(db.Sample{
							Host:        measurement.Host(),
							CountryCode: ps.s.CountryCode,
							Token:       r.token,
							ASN:         ps.s.ASN,
							Type:        measurement.Type().String(),
							Origin:      sampleorigins.Central.String(),
							Data:        data,
						})
						if err != nil {
							lg.Errorln(err.Error())
							continue measurerLoop
						}
					default:
						lg.Errorf("could not measure:%v error:%s", v, err.Error())
						continue measurerLoop
					}
				}
			}
		}()
	}
}

func queueMeasurements(token shared.SuggestionToken, measurers ...measure.Measurer) {
	requestMeasurements <- centralMeasurer{
		token: token,
		measurers: measurers,
	}
}
