// Package traefik_geoip is a Traefik plugin for Maxmind GeoIP2.
package traefik_geoip

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/IncSW/geoip2"
)

var lookup LookupGeoIP
var debug bool

// ResetLookup reset lookup function.
func ResetLookup() {
	lookup = nil
}

// Config the plugin configuration.
type Config struct {
	DBPath     string   `json:"dbPath,omitempty"`
	Debug      bool     `json:"debug,omitempty"`
	ExcludeIPs []string `json:"excludeIPs,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		DBPath:     DefaultDBPath,
		Debug:      DefaultDebug,
		ExcludeIPs: []string{},
	}
}

// TraefikGeoIP a traefik geoip plugin.
type TraefikGeoIP struct {
	next       http.Handler
	name       string
	ExcludeIPs []*net.IPNet
}

// New created a new TraefikGeoIP plugin.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	debug = cfg.Debug

	if debug {
		log.Printf("[geoip] setting up plugin: config=%v", cfg)
	}

	if _, err := os.Stat(cfg.DBPath); err != nil {
		if debug {
			log.Printf("[geoip] DB not found: db=%s, name=%s, err=%v", cfg.DBPath, name, err)
		}
		return &TraefikGeoIP{
			next: next,
			name: name,
		}, nil
	}

	if lookup == nil && strings.Contains(cfg.DBPath, "City") {
		rdr, err := geoip2.NewCityReaderFromFile(cfg.DBPath)
		if err != nil {
			if debug {
				log.Printf("[geoip] lookup DB is not initialized: db=%s, name=%s, err=%v", cfg.DBPath, name, err)
			}
		} else {
			lookup = CreateCityDBLookup(rdr)
			if debug {
				log.Printf("[geoip] lookup DB initialized: db=%s, name=%s, lookup=%v", cfg.DBPath, name, lookup)
			}
		}
	}

	if lookup == nil && strings.Contains(cfg.DBPath, "Country") {
		rdr, err := geoip2.NewCountryReaderFromFile(cfg.DBPath)
		if err != nil {
			if debug {
				log.Printf("[geoip] lookup DB is not initialized: db=%s, name=%s, err=%v", cfg.DBPath, name, err)
			}
		} else {
			lookup = CreateCountryDBLookup(rdr)
			if debug {
				log.Printf("[geoip] lookup DB initialized: db=%s, name=%s, lookup=%v", cfg.DBPath, name, lookup)
			}
		}
	}

	// Parse CIDRs and store them in a slice for exclusion.
	excludedIPs := []*net.IPNet{}
	for _, v := range cfg.ExcludeIPs {
		// Check if it is a single IP.
		if net.ParseIP(v) != nil {
			// Make the IP into a /32.
			v += "/32"
		}
		// Now parse the value as CIDR.
		_, excludedNet, err := net.ParseCIDR(v)
		if err != nil {
			// Ignore invalid CIDRs and continue.
			if debug {
				log.Printf("[geoip] invalid CIDR: cidr=%s, name=%s, err=%v", v, name, err)
			}
			continue
		}

		excludedIPs = append(excludedIPs, excludedNet)
	}

	return &TraefikGeoIP{
		next:       next,
		name:       name,
		ExcludeIPs: excludedIPs,
	}, nil
}

// isExcluded checks if the IP is in the exclude list.
func (mw *TraefikGeoIP) isExcluded(ip net.IP) bool {
	for _, net := range mw.ExcludeIPs {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}

// ProcessRequest processes the request and adds geo headers if the IP is in the database.
func (mw *TraefikGeoIP) ProcessRequest(req *http.Request) *http.Request {
	// Only process if the plugin was initialized
	if lookup == nil {
		if debug {
			log.Printf("[geoip] lookup is not initialized: name=%s", mw.name)
		}
		return req
	}

	// Get the remote IP and parse the host if needed.
	ipStr := req.RemoteAddr
	host, _, err := net.SplitHostPort(ipStr)
	if err == nil {
		ipStr = host
	}
	ip := net.ParseIP(ipStr)

	// Only process IPs not in the exclude list.
	if mw.isExcluded(ip) {
		if debug {
			log.Printf("[geoip] IP excluded: ip=%s, name=%s", ipStr, mw.name)
		}
		return req
	}

	// Lookup the IP.
	result, err := lookup(ip)
	if err != nil {
		if debug {
			log.Printf("[geoip] lookup error: ip=%s, name=%s, err=%v", ipStr, mw.name, err)
		}
		return req
	}

	if debug {
		log.Printf("[geoip] lookup result: ip=%s, name=%s, result=%v", ipStr, mw.name, result)
	}

	// Add the headers we have data for.
	if result.country != Unknown {
		req.Header.Set(CountryHeader, result.country)
	}
	if result.countryCode != Unknown {
		req.Header.Set(CountryCodeHeader, result.countryCode)
	}
	if result.region != Unknown {
		req.Header.Set(RegionHeader, result.region)
	}
	if result.city != Unknown {
		req.Header.Set(CityHeader, result.city)
	}
	if result.latitude != Unknown {
		req.Header.Set(LatitudeHeader, result.latitude)
	}
	if result.longitude != Unknown {
		req.Header.Set(LongitudeHeader, result.longitude)
	}

	return req
}

// ServeHTTP implements the middleware interface.
func (mw *TraefikGeoIP) ServeHTTP(reqWr http.ResponseWriter, req *http.Request) {
	// Process the request.
	req = mw.ProcessRequest(req)

	mw.next.ServeHTTP(reqWr, req)
}
