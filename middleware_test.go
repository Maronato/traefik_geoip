package traefikgeoip_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/Maronato/traefik_geoip"
)

const (
	ValidIP       = "188.193.88.199"
	ValidIPNoCity = "20.1.184.61"
)

func TestGeoIPConfig(t *testing.T) {
	mwCfg := mw.CreateConfig()
	if mw.DefaultDBPath != mwCfg.DBPath {
		t.Fatalf("Incorrect path")
	}

	mwCfg.DBPath = "./non-existing"
	mw.ResetLookup()
	_, err := mw.New(context.TODO(), nil, mwCfg, "")
	if err != nil {
		t.Fatalf("Must not fail on missing DB")
	}

	mwCfg.DBPath = "justfile"
	_, err = mw.New(context.TODO(), nil, mwCfg, "")
	if err != nil {
		t.Fatalf("Must not fail on invalid DB format")
	}
}

func TestGeoIPBasic(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./GeoLite2-City.mmdb"

	called := false
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { called = true })

	mw.ResetLookup()
	instance, err := mw.New(context.TODO(), next, mwCfg, "traefik-")
	if err != nil {
		t.Fatalf("Error creating %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)

	instance.ServeHTTP(recorder, req)
	if recorder.Result().StatusCode != http.StatusOK {
		t.Fatalf("Invalid return code")
	}
	if called != true {
		t.Fatalf("next handler was not called")
	}
}

func TestMissingGeoIPDB(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./missing"

	called := false
	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) { called = true })

	mw.ResetLookup()
	instance, err := mw.New(context.TODO(), next, mwCfg, "traefik-")
	if err != nil {
		t.Fatalf("Error creating %v", err)
	}

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = "1.2.3.4"

	instance.ServeHTTP(recorder, req)
	if recorder.Result().StatusCode != http.StatusOK {
		t.Fatalf("Invalid return code")
	}
	if called != true {
		t.Fatalf("next handler was not called")
	}
	assertHeader(t, req, mw.CountryHeader, "")
	assertHeader(t, req, mw.CountryCodeHeader, "")
	assertHeader(t, req, mw.RegionHeader, "")
	assertHeader(t, req, mw.CityHeader, "")
	assertHeader(t, req, mw.LatitudeHeader, "")
	assertHeader(t, req, mw.LongitudeHeader, "")
}

func TestGeoIPFromRemoteAddr(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./GeoLite2-City.mmdb"
	mwCfg.Debug = true

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	mw.ResetLookup()
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIP)
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "Germany")
	assertHeader(t, req, mw.CountryCodeHeader, "DE")
	assertHeader(t, req, mw.RegionHeader, "BY")
	assertHeader(t, req, mw.CityHeader, "Munich")
	assertHeader(t, req, mw.LatitudeHeader, "48.1663")
	assertHeader(t, req, mw.LongitudeHeader, "11.5683")

	req = httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIPNoCity)
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "United States")
	assertHeader(t, req, mw.CountryCodeHeader, "US")
	assertHeader(t, req, mw.RegionHeader, "")
	assertHeader(t, req, mw.CityHeader, "")
	assertHeader(t, req, mw.LatitudeHeader, "37.751")
	assertHeader(t, req, mw.LongitudeHeader, "-97.822")

	req = httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = "qwerty:9999"
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "")
	assertHeader(t, req, mw.CountryCodeHeader, "")
	assertHeader(t, req, mw.RegionHeader, "")
	assertHeader(t, req, mw.CityHeader, "")
	assertHeader(t, req, mw.LatitudeHeader, "")
	assertHeader(t, req, mw.LongitudeHeader, "")
}

func TestGeoIPCountryDBFromRemoteAddr(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./GeoLite2-Country.mmdb"

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	mw.ResetLookup()
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIP)
	instance.ServeHTTP(httptest.NewRecorder(), req)

	assertHeader(t, req, mw.CountryHeader, "Germany")
	assertHeader(t, req, mw.CountryCodeHeader, "DE")
	assertHeader(t, req, mw.RegionHeader, "")
	assertHeader(t, req, mw.CityHeader, "")
	assertHeader(t, req, mw.LatitudeHeader, "")
	assertHeader(t, req, mw.LongitudeHeader, "")
}

func TestIgnoresExcludedIPs(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./GeoLite2-City.mmdb"
	mwCfg.Debug = true
	mwCfg.ExcludeIPs = []string{ValidIP}

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	mw.ResetLookup()
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIP)
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "")
	assertHeader(t, req, mw.CountryCodeHeader, "")
	assertHeader(t, req, mw.RegionHeader, "")
	assertHeader(t, req, mw.CityHeader, "")
	assertHeader(t, req, mw.LatitudeHeader, "")
	assertHeader(t, req, mw.LongitudeHeader, "")
}

func TestHandleInvalidExcludeIP(t *testing.T) {
	mwCfg := mw.CreateConfig()
	mwCfg.DBPath = "./GeoLite2-City.mmdb"
	mwCfg.Debug = true
	mwCfg.ExcludeIPs = []string{"invalid"}

	next := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {})
	mw.ResetLookup()
	instance, _ := mw.New(context.TODO(), next, mwCfg, "traefik_geoip")

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req.RemoteAddr = fmt.Sprintf("%s:9999", ValidIP)
	instance.ServeHTTP(httptest.NewRecorder(), req)
	assertHeader(t, req, mw.CountryHeader, "Germany")
	assertHeader(t, req, mw.CountryCodeHeader, "DE")
	assertHeader(t, req, mw.RegionHeader, "BY")
	assertHeader(t, req, mw.CityHeader, "Munich")
	assertHeader(t, req, mw.LatitudeHeader, "48.1663")
	assertHeader(t, req, mw.LongitudeHeader, "11.5683")
}

func assertHeader(t *testing.T, req *http.Request, key, expected string) {
	t.Helper()
	if req.Header.Get(key) != expected {
		t.Fatalf("invalid value of header [%s] is '%s', not '%s'", key, expected, req.Header.Get(key))
	}
}
