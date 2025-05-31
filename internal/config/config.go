package config

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"

	"github.com/Myles-J/chirpy/internal/database"
)

// APIConfig holds the configuration for the API.
type APIConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
	polkaSecret    string
}

// NewAPIConfig creates a new APIConfig instance.
func NewAPIConfig(db *database.Queries, platform string, jwtSecret string, polkaSecret string) *APIConfig {
	return &APIConfig{
		db:          db,
		platform:    platform,
		jwtSecret:   jwtSecret,
		polkaSecret: polkaSecret,
	}
}

// HandlerMetrics increments the fileserver hits counter.
func (cfg *APIConfig) HandlerMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// MetricsHandler returns the metrics page.
func (cfg *APIConfig) MetricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(fmt.Appendf(nil, `
	<html>
	<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
	</body>
	</html>
	`, cfg.fileserverHits.Load()))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}

// ResetHandler resets the fileserver hits counter.
// It requires the platform to be "dev" to be authorized.
func (cfg *APIConfig) ResetHandler(w http.ResponseWriter, _ *http.Request) {
	if cfg.platform != "dev" {
		http.Error(w, "Not authorized", http.StatusForbidden)
		return
	}
	err := cfg.db.Reset(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte("Hits reset to 0"))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
