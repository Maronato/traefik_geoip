package traefik_geoip //nolint:revive,stylecheck

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/IncSW/geoip2" //nolint:depguard
)

const (
	// Unknown constant for undefined data.
	Unknown = "XX"
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

// GeoIPResult in memory, this should have between 126 and 180 bytes. On average, consider 150 bytes.
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

// newCityDBLookup Create a new CityDBLookup.
func newCityDBLookup(rdr *geoip2.CityReader) LookupGeoIP {
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
			geohash:     EncodeGeoHash(rec.Location.Latitude, rec.Location.Longitude),
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

// newCountryDBLookup Create a new CountryDBLookup.
func newCountryDBLookup(rdr *geoip2.CountryReader) LookupGeoIP {
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

func newCacheWrapper(lookup LookupGeoIP, cacheSize int) LookupGeoIP {
	cache := NewCache(cacheSize)

	return func(ip net.IP) (*GeoIPResult, error) {
		if result, ok := cache.Get(ip.String()); ok {
			return &result, nil
		}

		result, err := lookup(ip)
		if err != nil {
			return nil, err
		}

		cache.Add(ip.String(), *result)

		return result, nil
	}
}

// NewLookup Create a new Lookup.
func NewLookup(dbPath string, cacheSize int) (LookupGeoIP, error) {
	var lookup LookupGeoIP

	switch {
	case strings.Contains(dbPath, "City"):
		rdr, err := geoip2.NewCityReaderFromFile(dbPath)
		if err != nil {
			return nil, err
		}
		lookup = newCityDBLookup(rdr)

	case strings.Contains(dbPath, "Country"):
		rdr, err := geoip2.NewCountryReaderFromFile(dbPath)
		if err != nil {
			return nil, err
		}
		lookup = newCountryDBLookup(rdr)

	default:
		return nil, fmt.Errorf("unable to parse Geo DB type: db=%s", dbPath)
	}

	cachedLookup := newCacheWrapper(lookup, cacheSize)

	return cachedLookup, nil
}
