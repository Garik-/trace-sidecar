package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
	serviceAddr  = "http://localhost:8000"
)

func initInternalServer(ctx context.Context, addr string) *http.Server {
	mux := http.NewServeMux()
	srv := newServer(ctx, addr)
	srv.Handler = mux
	mux.Handle("/metrics", promhttp.Handler())

	return srv
}

func initServer(ctx context.Context, addr string) *http.Server {
	srv := newServer(ctx, addr)

	target, _ := url.Parse(serviceAddr)
	proxy := httputil.NewSingleHostReverseProxy(target)

	handler := func(p *httputil.ReverseProxy) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println(r.URL)
			r.Host = target.Host
			p.ServeHTTP(w, r)
		})
	}

	// TODO: need custom middleware
	srv.Handler = otelhttp.NewHandler(handler(proxy), "sidecar")

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
