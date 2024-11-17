package main

import (
	"io"
	"log"
	"net/http"
)

func sidecarHandler(addr string, client *http.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Host = addr
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

		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("io copy error: %v\n", err)
		}
	})
}
