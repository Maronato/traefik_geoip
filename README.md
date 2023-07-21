# Traefik plugin for MaxMind GeoIP

This is a modified version of [GiGInnovationLabs/traefikgeoip2](https://github.com/GiGInnovationLabs/traefikgeoip2) that changes the following:
- Logs are hidden by default, and can be displayed by setting the `debug: true` config
- It adds latitude, longitude, geohash, and the country name (moving the country code to CountryCode)
- It removes the `X-` from the header names, per [RFC 6648](https://www.rfc-editor.org/rfc/rfc6648).
- It adds support for `excludeIPs`, a config that takes IPs and CIDRs that will be excluded from checks
- It doesn't add a header if its value could not be determined
- To get the client's IP, it looks for it first in the `X-Forwarded-For` header and, if it's not there, it takes it from `req.remoteAddr`
- I had issues with Traefik not using the correct IP in `X-Real-IP`, so there's also a flag `setRealIP: true` that resets the header to the IP found in `X-Forwarded-For`.
- Added a cache layer to the lookup, based on [TinyLFU](https://github.com/dgryski/go-tinylfu) and controlled by `cacheSize: int`
---

[Traefik](https://doc.traefik.io/traefik/) plugin 
that registers a custom middleware 
for getting data from 
[MaxMind GeoIP databases](https://www.maxmind.com/en/geoip2-services-and-databases) 
and pass it downstream via HTTP request headers.

Supports both 
[GeoIP2](https://www.maxmind.com/en/geoip2-databases) 
and 
[GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) databases.

## Docs are TBD!
