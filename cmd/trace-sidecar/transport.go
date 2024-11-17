package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	readHeaderTimeout = 2 * time.Second
)

func newServer(ctx context.Context, addr string) *http.Server {
	return &http.Server{
		ReadHeaderTimeout: readHeaderTimeout,
		Addr:              addr,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
}

func listenAndServe(g *errgroup.Group, gCtx context.Context, server *http.Server) {
	g.Go(func() error {
		log.Printf("listening on %s\n", server.Addr)
		return server.ListenAndServe()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return server.Shutdown(context.Background())
	})
}
