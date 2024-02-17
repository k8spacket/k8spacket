package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/k8spacket/k8spacket/broker"
	"github.com/k8spacket/k8spacket/ebpf"
	k8spacket_log "github.com/k8spacket/k8spacket/log"
	"github.com/k8spacket/k8spacket/plugins"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	k8spacket_log.BuildLogger()

	pluginManager := plugins.NewPluginManager()
	plugins.InitPlugins(pluginManager)
	go broker.DistributeEvents(pluginManager)
	ebpf.LoadEbpf()
	handleEndpoints()
}

func handleEndpoints() {
	listenerPort := os.Getenv("K8S_PACKET_TCP_LISTENER_PORT")
	k8spacket_log.LOGGER.Printf("[api] Serving requests on port %s", listenerPort)
	prometheus.MustRegister(collectors.NewBuildInfoCollector())
	srv := &http.Server{Addr: fmt.Sprintf(":%s", listenerPort)}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			k8spacket_log.LOGGER.Fatalf("[api] Cannot start ListenAndServe: %+v", err)
		}

	}()

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	if err := srv.Shutdown(ctx); err != nil {
		k8spacket_log.LOGGER.Fatalf("[graceful] Server shutdown failed:%+v", err)
	}
	k8spacket_log.LOGGER.Print("[graceful] Application closed gracefully")
}
