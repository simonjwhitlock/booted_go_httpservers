package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	cfg.fileserverHits.Add(1)
	return next
}

func main() {
	var apiCfg apiConfig
	apiCfg.fileserverHits.Store(0)
	mux := http.NewServeMux()
	// Serve the logo.png file at the /assets path
	// Serve files from the current directory under the /app/ path, stripping the /app/ prefix
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	// Serve the logo.png file at the /assets path
	mux.Handle("/assets/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("/metrics/", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		str := fmt.Sprintf("Hits: %v", apiCfg.fileserverHits.Load())
		w.Write([]byte(str))
	})

	Server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	Server.ListenAndServe()

}
