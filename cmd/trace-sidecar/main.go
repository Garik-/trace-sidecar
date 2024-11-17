package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	_ "go.uber.org/automaxprocs"
	"golang.org/x/sync/errgroup"
)

const (
	defaultAddr  = "localhost:8001"
	internalAddr = "localhost:2223"
	serviceAddr  = "localhost:8000"
)

func initInternalServer(ctx context.Context, addr string) *http.Server {
	mux := http.NewServeMux()
	srv := newServer(ctx, addr)
	srv.Handler = mux
	mux.Handle("/metrics", promhttp.Handler())

	return srv
}

func initServer(ctx context.Context, addr string) *http.Server {
	client := newPooledClient()

	srv := newServer(ctx, addr)
	srv.Handler = otelhttp.NewHandler(sidecarHandler(serviceAddr, client), "sidecar")

	return srv
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	shutdownMeterProvider, err := initMeterProvider()
	if err != nil {
		log.Fatal(err)
	}

	defer shutdownMeterProvider()

	g, gCtx := errgroup.WithContext(ctx)

	listenAndServe(g, gCtx, initServer(ctx, defaultAddr))
	listenAndServe(g, gCtx, initInternalServer(ctx, internalAddr))

	if err := g.Wait(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
