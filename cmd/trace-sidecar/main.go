package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"

	_ "go.uber.org/automaxprocs"
	"golang.org/x/sync/errgroup"

	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

const (
	defaultHost = ""
	defaultPort = "8001"

	serviceAddr = "localhost:8000"
)

func initMeter() (*sdkmetric.MeterProvider, error) {
	exp, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exp)))
	otel.SetMeterProvider(mp)
	return mp, nil
}

// defaultPooledTransport returns a new http.Transport with similar default
// values to http.DefaultTransport. Do not use this for transient transports as
// it can leak file descriptors over time. Only use this for transports that
// will be re-used for the same host(s).
func defaultPooledTransport() *http.Transport {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   runtime.GOMAXPROCS(0) + 1,
	}
	return transport
}

// defaultPooledClient returns a new http.Client with similar default values to
// http.Client, but with a shared Transport. Do not use this function for
// transient clients as it can leak file descriptors over time. Only use this
// for clients that will be re-used for the same host(s).
func defaultPooledClient() *http.Client {
	return &http.Client{
		Transport: defaultPooledTransport(),
	}
}

// copyHeader implementation is taken from http.Header.Clone()
func copyHeader(h, h2 http.Header) {
	if h == nil {
		return
	}

	// Find total number of values.
	nv := 0
	for _, vv := range h {
		nv += len(vv)
	}
	sv := make([]string, nv) // shared backing array for headers' values
	for k, vv := range h {
		if vv == nil {
			// Preserve nil values. ReverseProxy distinguishes
			// between nil and zero-length header values.
			h2[k] = nil
			continue
		}
		n := copy(sv, vv)
		h2[k] = sv[:n:n]
		sv = sv[n:]
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	mp, err := initMeter()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := mp.Shutdown(context.Background()); err != nil {
			log.Printf("error shutting down meter provider: %v", err)
		}
	}()

	client := defaultPooledClient()

	httpServer := &http.Server{
		Addr: net.JoinHostPort(defaultHost, defaultPort),
		Handler: otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Host = serviceAddr
			r.RequestURI = ""
			r.URL.Scheme = "http"
			r.URL.Host = r.Host

			resp, err := client.Do(r)
			if err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				log.Printf("error client request: %v\n", err)
				return
			}
			defer resp.Body.Close()

			copyHeader(w.Header(), resp.Header)
			w.WriteHeader(resp.StatusCode)

			io.Copy(w, resp.Body)

			// if _, err := io.Copy(w, resp.Body); err != nil {
			// 	log.Printf("io copy error: %v\n", err)
			// }
		}), "sidecar"),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		log.Printf("listening on %s\n", httpServer.Addr)
		return httpServer.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return httpServer.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
