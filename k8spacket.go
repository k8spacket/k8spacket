package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/k8spacket/k8spacket/broker"
	"github.com/k8spacket/k8spacket/ebpf"
	"github.com/k8spacket/k8spacket/modules/nodegraph"
	tlsparser "github.com/k8spacket/k8spacket/modules/tls-parser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	nodegraphListener := nodegraph.Init()
	tlsParserListener := tlsparser.Init()
	b := broker.Init(nodegraphListener, tlsParserListener)

	go b.DistributeEvents()
	ebpf.Init(b)

	prometheus.MustRegister(collectors.NewBuildInfoCollector())

	startHttpServer()
}

func startHttpServer() {
	listenerPort := os.Getenv("K8S_PACKET_TCP_LISTENER_PORT")
	slog.Info("[api] Serving requests", "Port", listenerPort)

	srv := &http.Server{Addr: fmt.Sprintf(":%s", listenerPort)}
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			slog.Error("[api] Cannot start ListenAndServe", "Error", err)
		}

	}()

	// graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("[graceful] Server shutdown failed", "Error", err)
	}
	slog.Info("[graceful] Application closed gracefully")
}
