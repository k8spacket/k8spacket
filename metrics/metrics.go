package metrics

import (
	"github.com/k8spacket/k8s"
	"github.com/k8spacket/metrics/nodegraph"
	"github.com/k8spacket/metrics/prometheus"
	"github.com/likexian/whois"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
)

var reverseLookupMap = make(map[string]string)

func PushK8sPacketMetric(src string, srcPort string, dst string, dstPort string, closed bool, bytesSent float64, bytesReceived float64, duration float64) {
	hideSrcPort, _ := strconv.ParseBool(os.Getenv("K8S_PACKET_HIDE_SRC_PORT"))
	var srcPortMetrics = srcPort
	if hideSrcPort {
		srcPortMetrics = "dynamic"
	}

	var srcName = k8s.K8sInfo[src].Name
	if srcName == "" {
		srcName = reverseLookup(src)
	}

	var dstName = k8s.K8sInfo[dst].Name
	if dstName == "" {
		dstName = reverseLookup(dst)
	}

	prometheus.K8sPacketBytesSentMetric.WithLabelValues(k8s.K8sInfo[src].Namespace, src, srcName, srcPortMetrics, dst, dstName, dstPort, strconv.FormatBool(closed)).Observe(bytesSent)
	prometheus.K8sPacketBytesReceivedMetric.WithLabelValues(k8s.K8sInfo[src].Namespace, src, srcName, srcPortMetrics, dst, dstName, dstPort, strconv.FormatBool(closed)).Observe(bytesReceived)
	prometheus.K8sPacketDurationSecondsMetric.WithLabelValues(k8s.K8sInfo[src].Namespace, src, srcName, srcPortMetrics, dst, dstName, dstPort, strconv.FormatBool(closed)).Observe(duration)

	nodegraph.UpdateNodeGraph(src, srcName, k8s.K8sInfo[src].Namespace, dst, dstName, k8s.K8sInfo[dst].Namespace, closed, bytesSent, bytesReceived, duration)

	log.Printf("Connection: src=%v srcName=%v srcPort=%v srcNS=%v dst=%v dstName=%v dstPort=%v dstNS=%v closed=%v bytesSent=%v bytesReceived=%v duration=%v",
		src, srcName, srcPort, k8s.K8sInfo[src].Namespace, dst, dstName, dstPort, k8s.K8sInfo[dst].Namespace, closed, bytesSent, bytesReceived, duration)
}

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
			reverseLookup += "(" + record.Country.IsoCode + ", " + record.City.Names["en"] + ")"
		}
		reverseLookupMap[ip] = reverseLookup
	}
	return reverseLookupMap[ip]
}

// Check if an ip is private.
func privateIPCheck(ip string) bool {
	ipAddress := net.ParseIP(ip)
	return ipAddress.IsPrivate()
}
