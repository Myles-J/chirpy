package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/Myles-J/chirpy/internal/api"
	"github.com/Myles-J/chirpy/internal/config"
	"github.com/Myles-J/chirpy/internal/database"
	"github.com/Myles-J/chirpy/internal/utils"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const readHeaderTimeout = 5 * time.Second

func main() {
	const port = "8080"

	// Load environment variables
	if err := godotenv.Load("../../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get required environment variables
	dbURL := utils.MustGetenv("DB_URL")
	platform := utils.MustGetenv("PLATFORM")
	jwtSecret := utils.MustGetenv("JWT_SECRET")
	polkaSecret := utils.MustGetenv("POLKA_SECRET")

	// Database setup
	dbConn, dbOpenErr := sql.Open("postgres", dbURL)
	if dbOpenErr != nil {
		log.Fatal("Error opening database:", dbOpenErr)
	}
	defer dbConn.Close()

	dbQueries := database.New(dbConn)
	apiCfg := config.NewAPIConfig(dbQueries, platform, jwtSecret, polkaSecret)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// --- Static and App Endpoints ---
	mux.Handle("/app/", apiCfg.HandlerMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})))
	mux.Handle(
		"/app/assets/",
		apiCfg.HandlerMetrics(http.StripPrefix("/app/assets/", http.FileServer(http.Dir("assets")))),
	)

	// --- Health Check Endpoint ---
	mux.HandleFunc("GET /api/healthz", api.HandleHealthCheck)

	// --- Admin Endpoints ---
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetHandler)

	// --- Authentication Endpoints ---
	mux.HandleFunc("POST /api/login", api.LoginHandler(dbQueries, jwtSecret))
	mux.HandleFunc("POST /api/refresh", api.RefreshHandler(dbQueries, jwtSecret))
	mux.HandleFunc("POST /api/revoke", api.RevokeHandler(dbQueries))

	// --- User Endpoints ---
	mux.HandleFunc("POST /api/users", api.CreateUserHandler(dbQueries))
	mux.HandleFunc("PUT /api/users", api.UpdateUserHandler(dbQueries, jwtSecret))

	// --- Chirp Endpoints ---
	mux.HandleFunc("POST /api/chirps", api.CreateChirpHandler(dbQueries, jwtSecret))
	mux.HandleFunc("DELETE /api/chirps/{id}", api.DeleteChirpHandler(dbQueries, jwtSecret))
	mux.HandleFunc("GET /api/chirps", api.ListChirpsHandler(dbQueries))
	mux.HandleFunc("GET /api/chirps/{id}", api.GetChirpHandler(dbQueries))

	// ---- Polka Endpoint ----
	mux.HandleFunc("POST /api/polka/webhooks", api.PolkaWebhookHandler(dbQueries, polkaSecret))

	// Server configuration and start
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           mux,
		ReadHeaderTimeout: readHeaderTimeout,
	}

	log.Printf("Starting server on port %s", port)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Server error: %v", err)
	}
}
