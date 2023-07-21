// Package traefik_geoip is a Traefik plugin for Maxmind GeoIP2.
package traefik_geoip //nolint:revive,stylecheck

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

const (
	// DefaultDBPath default GeoIP2 database path.
	DefaultDBPath = "GeoLite2-Country.mmdb"
	// defaultDebug default debug.
	defaultDebug = false
	// defaultSetRealIP default set real IP.
	defaultSetRealIP = false
	// defaultCacheSize default cache size.
	defaultCacheSize = 1000
)

// Config the plugin configuration.
type Config struct {
	DBPath     string   `json:"dbPath,omitempty"`
	Debug      bool     `json:"debug,omitempty"`
	ExcludeIPs []string `json:"excludeIPs,omitempty"`
	SetRealIP  bool     `json:"setRealIP,omitempty"` //nolint:tagliatelle
	CacheSize  int      `json:"cacheSize,omitempty"`
}

// CreateConfig creates the default plugin configuration.
func CreateConfig() *Config {
	return &Config{
		DBPath:     DefaultDBPath,
		Debug:      defaultDebug,
		ExcludeIPs: []string{},
		SetRealIP:  defaultSetRealIP,
		CacheSize:  defaultCacheSize,
	}
}

// TraefikGeoIP a traefik geoip plugin.
type TraefikGeoIP struct {
	next       http.Handler
	name       string
	excludeIPs []*net.IPNet
	lookup     LookupGeoIP
	debug      bool
	setRealIP  bool
}

// New created a new TraefikGeoIP plugin.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	debug := cfg.Debug

	if debug {
		log.Printf("[geoip] setting up plugin: config=%v", cfg)
	}

	if _, err := os.Stat(cfg.DBPath); err != nil {
		return nil, err
	}

	// Initialize the lookup DB.
	lookup, err := NewLookup(cfg.DBPath, cfg.CacheSize)
	if err != nil {
		if debug {
			log.Printf("[geoip] error initializing lookup: err=%v", err)
		}
		return nil, err
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
		excludeIPs: excludedIPs,
		lookup:     lookup,
		debug:      debug,
		setRealIP:  cfg.SetRealIP,
	}, nil
}

// isExcluded checks if the IP is in the exclude list.
func (mw *TraefikGeoIP) isExcluded(ip net.IP) bool {
	for _, net := range mw.excludeIPs {
		if net.Contains(ip) {
			return true
		}
	}

	return false
}

func (mw *TraefikGeoIP) getClientIP(req *http.Request) net.IP {
	// Get first IP from X-Forwarded-For header if it exists.
	ipStr := ""
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ipStr = strings.TrimSpace(ips[0])
		}
	}

	// If X-Forwarded-For header is empty, get the remote IP.
	if ipStr == "" {
		ipStr = req.RemoteAddr
		host, _, err := net.SplitHostPort(ipStr)
		if err == nil {
			ipStr = host
		}
	}

	// Parse the IP.
	ip := net.ParseIP(ipStr)
	if ip == nil && mw.debug {
		log.Printf("[geoip] unable to parse IP: ip=%s, name=%s", ipStr, mw.name)
		return ip
	}

	// Only process IPs not in the exclude list.
	if mw.isExcluded(ip) {
		if mw.debug {
			log.Printf("[geoip] IP excluded: ip=%s, name=%s", ipStr, mw.name)
		}
		ip = nil
	}

	return ip
}

// processRequest processes the request and adds geo headers if the IP is in the database.
func (mw *TraefikGeoIP) processRequest(req *http.Request) *http.Request {
	// Get the client IP.
	ip := mw.getClientIP(req)

	// If the IP is nil, return the request unchanged.
	if ip == nil {
		return req
	}

	// Set X-Real-Ip header because traefik sometimes messes with it.
	if mw.setRealIP {
		req.Header.Set("X-Real-Ip", ip.String())
	}

	// Lookup the IP.
	result, err := mw.lookup(ip)
	if err != nil {
		if mw.debug {
			log.Printf("[geoip] lookup error: ip=%v, name=%s, err=%v", ip, mw.name, err)
		}
		return req
	}

	if mw.debug {
		log.Printf("[geoip] lookup result: ip=%v, name=%s, result=%v", ip, mw.name, result)
	}

	// Set the headers.
	setHeaders(req, result)

	return req
}

// ServeHTTP implements the middleware interface.
func (mw *TraefikGeoIP) ServeHTTP(reqWr http.ResponseWriter, req *http.Request) {
	req = mw.processRequest(req)

	mw.next.ServeHTTP(reqWr, req)
}

// SetHeaders Set geo headers.
func setHeaders(req *http.Request, result *GeoIPResult) {
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
	if result.geohash != Unknown {
		req.Header.Set(GeohashHeader, result.geohash)
	}
}
