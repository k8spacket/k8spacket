package ebpf_tools

import (
	"github.com/k8spacket/k8s-api/v2"
	"github.com/k8spacket/k8spacket/modules"
	"github.com/likexian/whois"
	"github.com/oschwald/geoip2-golang"
	"net"
	"os"
	"regexp"
)

var reverseLookupMap = make(map[string]string)

var K8sInfo = make(map[string]k8s.IPResourceInfo)

func EnrichAddress(addr *modules.Address) {
	addr.Name = K8sInfo[addr.Addr].Name
	if addr.Name == "" {
		addr.Name = reverseLookup(addr.Addr)
	}
	addr.Namespace = K8sInfo[addr.Addr].Namespace
}

// try to find organization name and (if GeoLite2 Free Geolocation Data enabled) country and city by external IP
func reverseLookup(ip string) string {

	if privateIPCheck(ip) {
		return "N/A"
	}

	if _, ok := reverseLookupMap[ip]; !ok {

		result, _ := whois.Whois(ip)

		re := regexp.MustCompile(os.Getenv("K8S_PACKET_REVERSE_WHOIS_REGEXP"))
		matches := re.FindStringSubmatch(result)

		reverseLookup := ""

		if len(matches) > 1 {
			reverseLookup += matches[1]
		}

		db, err := geoip2.Open(os.Getenv("K8S_PACKET_REVERSE_GEOIP2_DB_PATH"))
		if err == nil {
			defer db.Close()

			ipObj := net.ParseIP(ip)
			record, _ := db.City(ipObj)
			if len(record.Country.IsoCode) > 0 && len(record.City.Names["en"]) > 0 {
				reverseLookup += "(" + record.Country.IsoCode + ", " + record.City.Names["en"] + ")"
			}
		}
		reverseLookupMap[ip] = reverseLookup
	}
	return reverseLookupMap[ip]
}

// Check if an IP is private.
func privateIPCheck(ip string) bool {
	ipAddress := net.ParseIP(ip)
	return ipAddress.IsPrivate()
}

func SliceContains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
