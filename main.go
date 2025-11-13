package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/simonjwhitlock/booted_go_httpservers/internal/database"
)

type apiConfig struct {
	fileserverHits int
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func main() {
	// load DB connection string from .env and then setup db connection
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("error setting up database connection: %v", err)
		return
	}

	const filepathRoot = "."
	const port = "8080"

	apiCfg := &apiConfig{
		fileserverHits: 0,
		dbQueries:      database.New(db),
	}

	mux := http.NewServeMux()
	// Serve the logo.png file at the /assets path
	// Serve files from the current directory under the /app/ path, stripping the /app/ prefix
	handler := http.FileServer(http.Dir(filepathRoot))
	mux.Handle("/app/", http.StripPrefix("/app/", apiCfg.middlewareMetricsInc(handler)))
	// Serve the logo.png file at the /assets path
	mux.Handle("/api/assets", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("GET /api/healthz", http.HandlerFunc(handlerReadiness))
	mux.Handle("GET /admin/metrics", http.HandlerFunc(apiCfg.handlerMetrics))
	mux.Handle("POST /admin/reset", http.HandlerFunc(apiCfg.handlerResetDEV))
	mux.Handle("POST /api/chirps", http.HandlerFunc(apiCfg.handlerChirps))
	mux.Handle("GET /api/chirps", http.HandlerFunc(apiCfg.handlerGetChrips))
	mux.Handle("GET /api/chirps/{chirpID}", http.HandlerFunc(apiCfg.handlerGetChirp))
	mux.Handle("POST /api/users", http.HandlerFunc(apiCfg.handlerUserRegistration))
	mux.Handle("POST /api/login", http.HandlerFunc(apiCfg.handlerUserLogin))

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
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	str := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", c.fileserverHits)
	w.Write([]byte(str))
}

func (c *apiConfig) handlerResetDEV(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if os.Getenv("PLATFORM") == "dev" {
		err := c.dbQueries.ResetUsers(req.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error resetting users: %v", err)))
		} else {
			c.fileserverHits = 0
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Platform reset"))
		}
	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("site not in dev mode"))
	}
}
