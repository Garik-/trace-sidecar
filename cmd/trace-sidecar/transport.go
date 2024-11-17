package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"golang.org/x/sync/errgroup"
)

func newServer(ctx context.Context, addr string) *http.Server {
	return &http.Server{
		Addr: addr,
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
