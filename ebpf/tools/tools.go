package ebpf_tools

import (
	"fmt"
	k8sclient "github.com/k8spacket/k8spacket/external/k8s"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/k8spacket/k8spacket/modules"
	"github.com/likexian/whois"
	"github.com/oschwald/geoip2-golang"
)

const (
	id_format string = "%s-%d"
)

var domainsMap = make(map[string]string)
var reverseLookupMap = make(map[string]string)

func EnrichAddress(addr *modules.Address) {
	name, namespace := k8sclient.GetNameAndNamespace(addr.Addr)
	addr.Name = name
	if addr.Name == "" {
		addr.Name = reverseLookup(addr.Addr, addr.Port)
	}
	addr.Namespace = namespace
}

// try to find domain (https only), organization name and (if GeoLite2 Free Geolocation Data enabled) country and city by external IP
func reverseLookup(ip string, port uint16) string {

	if privateIPCheck(ip) {
		return "N/A"
	}

	var name []string
	if val, ok := domainsMap[fmt.Sprintf(id_format, ip, port)]; ok {
		name = append(name, val)
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
	name = append(name, reverseLookupMap[ip])
	return strings.Join(name, ", ")
}

// Check if an IP is private.
func privateIPCheck(ip string) bool {
	ipAddress := net.ParseIP(ip)
	return ipAddress.IsPrivate()
}

func StoreDomain(ip string, port uint16, domain string) {
	if len(domain) > 0 {
		domainsMap[fmt.Sprintf(id_format, ip, port)] = domain
	}
}
