package main

import (
	"fmt"
	"log"
	"net/http"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func main() {

	const filepathRoot = "."
	const port = "8080"

	apiCfg := &apiConfig{
		fileserverHits: 0,
	}

	mux := http.NewServeMux()
	// Serve the logo.png file at the /assets path
	// Serve files from the current directory under the /app/ path, stripping the /app/ prefix
	handler := http.FileServer(http.Dir(filepathRoot))
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(handler)))
	// Serve the logo.png file at the /assets path
	mux.Handle("/assets/", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/healthz/", http.HandlerFunc(handlerReadiness))
	mux.Handle("/metrics/", http.HandlerFunc(apiCfg.handlerMetrics))
	mux.Handle("/reset/", http.HandlerFunc(apiCfg.handlerResetMetrics))

	Server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(Server.ListenAndServe())
}

func handlerReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (c *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	str := fmt.Sprintf("Hits: %v\n", c.fileserverHits)
	w.Write([]byte(str))
}

func (c *apiConfig) handlerResetMetrics(w http.ResponseWriter, req *http.Request) {
	c.fileserverHits = 0
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
