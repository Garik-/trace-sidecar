package main

import (
	"context"
	"io"
	"log"
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs"
	"golang.org/x/sync/errgroup"
)

const (
	defaultHost        = ""
	defaultPort        = "8000"
	minRequestDuration = 200  // ms
	maxRequestDuration = 2000 // ms
)

func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	httpServer := &http.Server{
		Addr: net.JoinHostPort(defaultHost, defaultPort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			d := time.Duration(randRange(minRequestDuration, maxRequestDuration)) * time.Millisecond

			select {
			case <-r.Context().Done():
			case <-time.After(d):
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			if _, err := io.WriteString(w, "request duration "+d.String()); err != nil {
				log.Printf("write error: %v\n", err)
			}
		}),
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
		log.Fatalf("exit reason: %s\n", err)
	}
}
