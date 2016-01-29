package db

import (
	"net"

	"github.com/oschwald/maxminddb-golang"
	"github.com/thomasf/lg"
)

type MaxmindClient interface {
	IP2CityGeoNameID(IP net.IP) uint
	IP2CountryCode(IP net.IP) string
}

type Maxmind struct {
	mmCountryDB *maxminddb.Reader
	mmCityDB    *maxminddb.Reader
}

func NewMaxmindClient(countryDB, cityDB *maxminddb.Reader) *Maxmind {
	return &Maxmind{
		mmCountryDB: countryDB,
		mmCityDB:    cityDB,
	}
}

// IP2CountryCode resolves an IP address to ISO country code using an geoip
// database.
func (m *Maxmind) IP2CountryCode(IP net.IP) string {
	var record onlyCountry // Or any appropriate struct
	err := m.mmCountryDB.Lookup(IP, &record)
	if err != nil {
		lg.Warning(err)
	}
	return record.Country.IsoCode
}

func (m *Maxmind) IP2CityGeoNameID(IP net.IP) uint {

	var record onlyCity // Or any appropriate struct
	if m.mmCityDB != nil {
		err := m.mmCityDB.Lookup(IP, &record)
		if err != nil {
			lg.Warning(err)
		}
		return record.City.GeoNameID
	}
	return 0
}

type onlyCountry struct {
	Country struct {
		IsoCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

type onlyCity struct {
	City struct {
		GeoNameID uint `maxminddb:"geoname_id"`
	} `maxminddb:"city"`
}
