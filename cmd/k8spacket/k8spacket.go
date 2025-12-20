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

	"github.com/k8spacket/k8spacket/internal/broker"
	"github.com/k8spacket/k8spacket/internal/ebpf"
	ebpf_inet "github.com/k8spacket/k8spacket/internal/ebpf/inet"
	ebpf_socketfilter "github.com/k8spacket/k8spacket/internal/ebpf/socketfilter"
	ebpf_tc "github.com/k8spacket/k8spacket/internal/ebpf/tc"
	"github.com/k8spacket/k8spacket/internal/modules/nodegraph"
	"github.com/k8spacket/k8spacket/internal/modules/tlsparser"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	mux := http.NewServeMux()

	nodegraphListener := nodegraph.Init(mux)
	tlsParserListener := tlsparser.Init(mux)
	distributionBroker := broker.Init(nodegraphListener, tlsParserListener)

	inetEbpf := &ebpf_inet.EbpfInet{Broker: distributionBroker}
	tcEbpf := &ebpf_tc.EbpfTc{Broker: distributionBroker}
	socketFilterEbpf := &ebpf_socketfilter.EbpfSocketFilter{Broker: distributionBroker}
	loader := ebpf.Init(inetEbpf, tcEbpf, socketFilterEbpf)

	buildLogger()
	startApp(distributionBroker, loader, mux)
}

func startApp(broker broker.Broker, loader ebpf.Loader, mux *http.ServeMux) {
	go broker.DistributeEvents()
	loader.Load()

	prometheus.MustRegister(collectors.NewBuildInfoCollector())
	startHttpServer(mux)
}

func startHttpServer(mux *http.ServeMux) {
	listenerPort := os.Getenv("K8S_PACKET_TCP_LISTENER_PORT")
	slog.Info("[api] Serving requests", "Port", listenerPort)

	srv := &http.Server{Addr: fmt.Sprintf(":%s", listenerPort), Handler: mux}
	go func() {
		mux.Handle("/metrics", promhttp.Handler())
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

func buildLogger() {
	var l slog.Level
	err := l.UnmarshalText([]byte(os.Getenv("LOG_LEVEL")))
	if err != nil {
		l = slog.LevelInfo
	}
	slog.SetLogLoggerLevel(l)
}
