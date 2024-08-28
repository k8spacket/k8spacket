package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/k8spacket/k8spacket/broker"
	"github.com/k8spacket/k8spacket/ebpf"
	k8spacket_log "github.com/k8spacket/k8spacket/log"
	"github.com/k8spacket/k8spacket/modules/nodegraph"
	_ "github.com/k8spacket/k8spacket/modules/nodegraph"
	_ "github.com/k8spacket/k8spacket/modules/tls-parser"
	tls_parser "github.com/k8spacket/k8spacket/modules/tls-parser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	k8spacket_log.BuildLogger()

	nodegraphListener := nodegraph.Init()
	tlsParserListener := tls_parser.Init()
	b := &broker.Broker{NodegraphListener: nodegraphListener, TlsParserListener: tlsParserListener}

	go b.DistributeEvents()
	ebpf.LoadEbpf(b)

	prometheus.MustRegister(collectors.NewBuildInfoCollector())

	startHttpServer()
}

func startHttpServer() {
	listenerPort := os.Getenv("K8S_PACKET_TCP_LISTENER_PORT")
	k8spacket_log.LOGGER.Printf("[api] Serving requests on port %s", listenerPort)

	srv := &http.Server{Addr: fmt.Sprintf(":%s", listenerPort)}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			k8spacket_log.LOGGER.Fatalf("[api] Cannot start ListenAndServe", "Error", err)
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
