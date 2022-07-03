package main

import (
	"github.com/k8spacket/metrics/nodegraph"
	"github.com/k8spacket/tcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
)

func main() {
	tcp.StartListeners()
	log.Printf("Serving requests on port 8080")
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/connections", nodegraph.ConnectionHandler)
	http.HandleFunc("/api/graph/fields", nodegraph.NodeGraphFieldsHandler)
	http.HandleFunc("/api/graph/data", nodegraph.NodeGraphDataHandler)
	http.HandleFunc("/api/health", nodegraph.Health)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
