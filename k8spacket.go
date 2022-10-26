package main

import (
	"fmt"
	"github.com/k8spacket/k8spacket/broker"
	"github.com/k8spacket/k8spacket/plugins"
	"github.com/k8spacket/k8spacket/tcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

func main() {
	pluginManager := plugins.NewPluginManager()
	plugins.InitPlugins(pluginManager)
	go broker.DistributeMessages(pluginManager)
	tcp.StartListeners()
	listenerPort := os.Getenv("K8S_PACKET_TCP_LISTENER_PORT")
	log.Printf("Serving requests on port %s", listenerPort)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", listenerPort), nil))
}
