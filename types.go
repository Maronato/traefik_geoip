package traefik_geoip

import (
	"fmt"
	"net"
	"strconv"

	"github.com/IncSW/geoip2"
)

// Unknown constant for undefined data.
const Unknown = "XX"

// DefaultDBPath default GeoIP2 database path.
const DefaultDBPath = "GeoLite2-Country.mmdb"

// DefaultDebug default debug.
const DefaultDebug = false

const (
	// CountryHeader country header name.
	CountryHeader = "GeoIP-Country"
	// CountryCodeHeader country code header name.
	CountryCodeHeader = "GeoIP-Country-Code"
	// RegionHeader region header name.
	RegionHeader = "GeoIP-Region"
	// CityHeader city header name.
	CityHeader = "GeoIP-City"
	// LatitudeHeader latitude header name.
	LatitudeHeader = "GeoIP-Latitude"
	// LongitudeHeader longitude header name.
	LongitudeHeader = "GeoIP-Longitude"
	// GeohashHeader geohash header name.
	GeohashHeader = "GeoIP-Geohash"
)

// GeoIPResult GeoIPResult.
type GeoIPResult struct {
	country     string
	countryCode string
	region      string
	city        string
	latitude    string
	longitude   string
	geohash     string
}

// LookupGeoIP LookupGeoIP.
type LookupGeoIP func(ip net.IP) (*GeoIPResult, error)

// CreateCityDBLookup CreateCityDBLookup.
func CreateCityDBLookup(rdr *geoip2.CityReader) LookupGeoIP {
	return func(ip net.IP) (*GeoIPResult, error) {
		rec, err := rdr.Lookup(ip)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		retval := GeoIPResult{
			country:     Unknown,
			countryCode: rec.Country.ISOCode,
			region:      Unknown,
			city:        Unknown,
			latitude:    strconv.FormatFloat(rec.Location.Latitude, 'f', -1, 64),
			longitude:   strconv.FormatFloat(rec.Location.Longitude, 'f', -1, 64),
			geohash:     Encode(rec.Location.Latitude, rec.Location.Longitude),
		}
		if country, ok := rec.Country.Names["en"]; ok {
			retval.country = country
		}
		if city, ok := rec.City.Names["en"]; ok {
			retval.city = city
		}
		if rec.Subdivisions != nil {
			retval.region = rec.Subdivisions[0].ISOCode
		}
		return &retval, nil
	}
}

// CreateCountryDBLookup CreateCountryDBLookup.
func CreateCountryDBLookup(rdr *geoip2.CountryReader) LookupGeoIP {
	return func(ip net.IP) (*GeoIPResult, error) {
		rec, err := rdr.Lookup(ip)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		retval := GeoIPResult{
			country:     Unknown,
			countryCode: rec.Country.ISOCode,
			region:      Unknown,
			city:        Unknown,
			latitude:    Unknown,
			longitude:   Unknown,
			geohash:     Unknown,
		}
		if country, ok := rec.Country.Names["en"]; ok {
			retval.country = country
		}
		return &retval, nil
	}
}
