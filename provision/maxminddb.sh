echo install geolite2

mkdir -p ${HOME}/.alkasir-central/internet
cd ${HOME}/.alkasir-central/internet
curl http://geolite.maxmind.com/download/geoip/database/GeoLite2-Country.mmdb.gz -o GeoLite2-Country.mmdb.gz
gunzip GeoLite2-Country.mmdb.gz
