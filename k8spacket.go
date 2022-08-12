package main

import (
	"github.com/k8spacket/metrics/nodegraph"
	"github.com/k8spacket/tcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func main() {
	var port = getEnv("HTTP_LISTENING_PORT", "8080")
	var listeningAddressPort = os.Getenv("HTTP_LISTENING_ADDR") + ":" + port
	tcp.StartListeners()
	log.Printf("Serving requests on " + listeningAddressPort)
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/connections", nodegraph.ConnectionHandler)
	http.HandleFunc("/api/graph/fields", nodegraph.NodeGraphFieldsHandler)
	http.HandleFunc("/api/graph/data", nodegraph.NodeGraphDataHandler)
	http.HandleFunc("/api/health", nodegraph.Health)
	log.Fatal(http.ListenAndServe(listeningAddressPort, nil))
}
