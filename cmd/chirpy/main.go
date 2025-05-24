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

	// Database setup
	dbConn, dbOpenErr := sql.Open("postgres", dbURL)
	if dbOpenErr != nil {
		log.Fatal("Error opening database:", dbOpenErr)
	}
	defer dbConn.Close()

	dbQueries := database.New(dbConn)
	apiCfg := config.NewAPIConfig(dbQueries, platform, jwtSecret)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register endpoints
	mux.Handle("/app/", apiCfg.HandlerMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})))
	mux.HandleFunc("GET /api/healthz", api.HandleHealthCheck)

	// Serve static files from the assets directory
	mux.Handle(
		"/app/assets/",
		apiCfg.HandlerMetrics(http.StripPrefix("/app/assets/", http.FileServer(http.Dir("assets")))),
	)

	// Register admin endpoints
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetHandler)

	// Register API endpoints
	mux.HandleFunc("POST /api/login", api.LoginHandler(dbQueries, jwtSecret))
	mux.HandleFunc("POST /api/refresh", api.RefreshHandler(dbQueries, jwtSecret))
	mux.HandleFunc("POST /api/revoke", api.RevokeHandler(dbQueries))
	mux.HandleFunc("POST /api/chirps", api.CreateChirpHandler(dbQueries, jwtSecret))
	mux.HandleFunc("POST /api/users", api.CreateUserHandler(dbQueries))
	mux.HandleFunc("GET /api/chirps", api.ListChirpsHandler(dbQueries))
	mux.HandleFunc("GET /api/chirps/{id}", api.GetChirpHandler(dbQueries))

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
