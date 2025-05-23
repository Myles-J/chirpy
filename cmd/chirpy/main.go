package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Myles-J/chirpy/internal/api"
	"github.com/Myles-J/chirpy/internal/config"
	"github.com/Myles-J/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	const port = "8080"
	// Load environment variables
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Database setup
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("DB_URL is not set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM is not set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is not set")
	}

	dbConn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Error opening database", err)
	}
	defer dbConn.Close()

	dbQueries := database.New(dbConn)
	apiCfg := config.NewApiConfig(dbQueries, platform, jwtSecret)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Define handler functions
	handleApp := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Register endpoints
	mux.Handle("/app/", apiCfg.HandlerMetrics(handleApp))
	mux.HandleFunc("GET /api/healthz", api.HandleHealthCheck)

	// Serve static files from the assets directory
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/app/assets/", apiCfg.HandlerMetrics(http.StripPrefix("/app/assets/", fs)))

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

	// Server configuration
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start the server
	log.Println("Starting server on port", port)
	log.Fatal(server.ListenAndServe())
}
