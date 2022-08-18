package main

import (
	"fmt"
	"github.com/k8spacket/metrics/nodegraph"
	"github.com/k8spacket/tcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

func main() {
	tcp.StartListeners()
	listenerPort := os.Getenv("K8S_PACKET_TCP_LISTENER_PORT")
	log.Printf("Serving requests on port %s", listenerPort)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/connections", nodegraph.ConnectionHandler)
	http.HandleFunc("/api/graph/fields", nodegraph.NodeGraphFieldsHandler)
	http.HandleFunc("/api/graph/data", nodegraph.NodeGraphDataHandler)
	http.HandleFunc("/api/health", nodegraph.Health)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", listenerPort), nil))
}
