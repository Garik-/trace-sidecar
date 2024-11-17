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
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	_ "go.uber.org/automaxprocs"
	"golang.org/x/sync/errgroup"
)

func statusOK(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
func statusOKHandler() http.Handler                   { return http.HandlerFunc(statusOK) }

func initInternalServer(ctx context.Context, addr string) *http.Server {
	mux := http.NewServeMux()
	srv := newServer(ctx, addr)
	srv.Handler = mux

	mux.Handle("/metrics", promhttp.Handler())
	mux.Handle("/livez", statusOKHandler())
	mux.Handle("/readyz", statusOKHandler())

	return srv
}

func initServer(ctx context.Context, cfg *config) *http.Server {
	srv := newServer(ctx, cfg.addr)

	target, _ := url.Parse(cfg.targetURL)
	proxy := httputil.NewSingleHostReverseProxy(target)

	handler := func(p *httputil.ReverseProxy) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			attr := semconv.HTTPRoute(r.URL.Path)
			log.Println(attr.Value.AsString())

			labeler, _ := otelhttp.LabelerFromContext(r.Context())
			labeler.Add(attr)

			span := trace.SpanFromContext(r.Context())
			span.SetAttributes(attr)

			r.Host = target.Host
			p.ServeHTTP(w, r)
		})
	}

	srv.Handler = otelhttp.NewHandler(handler(proxy), cfg.service.name)

	return srv
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := newConfig()
	log.Printf("config: %+v\n", cfg)

	shutdownMeterProvider, err := initMeterProvider(cfg.service)

	if err != nil {
		log.Println(err)
		return
	}

	defer shutdownMeterProvider()

	g, gCtx := errgroup.WithContext(ctx)

	listenAndServe(g, gCtx, initServer(ctx, cfg))
	listenAndServe(g, gCtx, initInternalServer(ctx, cfg.internalAddr))

	if err := g.Wait(); err != nil && err != http.ErrServerClosed {
		log.Println(err)
	}
}
