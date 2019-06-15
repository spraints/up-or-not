package main

import (
	"context"
	"log"
	"net/http"
)

func buildHTTPHandler(m *model) http.Handler {
	return http.NewServeMux()
	//mux.HandleFunc("/api/status", api.status)
	// TODO - add a handler
	// /api/status - {"status": "up|down"}
	// /           - web page that renders the above.
}

func serveHTTP(ctx context.Context, server *http.Server) error {
	log.Printf("HTTP server started (%s)", server.Addr)
	defer log.Printf("HTTP server stopped (%s)", server.Addr)

	done := make(chan error)
	go func() {
		defer close(done)
		done <- server.ListenAndServe()
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		log.Printf("Shutdown HTTP server")
		return server.Shutdown(ctx)
	}
}
