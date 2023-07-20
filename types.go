package traefikgeoip2

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

// Debug default
const DefaultDebug = false

const (
	// CountryHeader country header name.
	CountryHeader = "GeoIP-Country"
	// CountryHeader country code header name.
	CountryCodeHeader = "GeoIP-Country-Code"
	// RegionHeader region header name.
	RegionHeader = "GeoIP-Region"
	// CityHeader city header name.
	CityHeader = "GeoIP-City"
	// LatitudeHeader latitude header name.
	LatitudeHeader = "GeoIP-Latitude"
	// LongitudeHeader longitude header name.
	LongitudeHeader = "GeoIP-Longitude"
)

// GeoIPResult GeoIPResult.
type GeoIPResult struct {
	country     string
	countryCode string
	region      string
	city        string
	latitude    string
	longitude   string
}

// LookupGeoIP2 LookupGeoIP2.
type LookupGeoIP2 func(ip net.IP) (*GeoIPResult, error)

// CreateCityDBLookup CreateCityDBLookup.
func CreateCityDBLookup(rdr *geoip2.CityReader) LookupGeoIP2 {
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
func CreateCountryDBLookup(rdr *geoip2.CountryReader) LookupGeoIP2 {
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
		}
		if country, ok := rec.Country.Names["en"]; ok {
			retval.country = country
		}
		return &retval, nil
	}
}
