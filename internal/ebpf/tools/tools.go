package ebpf_tools

import (
	"fmt"
	k8sclient "github.com/k8spacket/k8spacket/internal/infra/k8s"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/k8spacket/k8spacket/pkg/events"
	"github.com/likexian/whois"
	"github.com/oschwald/geoip2-golang"
)

const (
	id_format string = "%s-%d"
)

type SafeMap struct {
	mu   sync.RWMutex
	data map[string]string
}

var domainsMap = &SafeMap{data: make(map[string]string)}
var reverseLookupMap = &SafeMap{data: make(map[string]string)}

func EnrichAddress(addr *events.Address) {
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
	domainsMap.mu.RLock()
	if val, ok := domainsMap.data[fmt.Sprintf(id_format, ip, port)]; ok {
		name = append(name, val)
	}
	domainsMap.mu.RUnlock()

	reverseLookupMap.mu.Lock()
	if _, ok := reverseLookupMap.data[ip]; !ok {

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
		reverseLookupMap.data[ip] = reverseLookup
	}
	if val, ok := reverseLookupMap.data[ip]; ok {
		name = append(name, val)
	}
	reverseLookupMap.mu.Unlock()
	return strings.Join(name, ", ")
}

// Check if an IP is private.
func privateIPCheck(ip string) bool {
	ipAddress := net.ParseIP(ip)
	return ipAddress.IsPrivate()
}

func StoreDomain(ip string, port uint16, domain string) {
	if len(domain) > 0 {
		domainsMap.mu.Lock()
		domainsMap.data[fmt.Sprintf(id_format, ip, port)] = domain
		domainsMap.mu.Unlock()
	}
}
