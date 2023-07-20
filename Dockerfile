FROM traefik:2.4.9

COPY GeoLite2-City.mmdb /var/lib/traefikgeoip/ 

