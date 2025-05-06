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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	dbUrl := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	dbConn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Error opening database", err)
	}
	defer dbConn.Close()

	dbQueries := database.New(dbConn)
	apiCfg := config.NewApiConfig(dbQueries, platform)

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.HandlerMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})))

	mux.HandleFunc("GET /api/healthz", api.HandleHealthCheck)

	// Serve static files from the assets directory
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/app/assets/", apiCfg.HandlerMetrics(http.StripPrefix("/app/assets/", fs)))

	// Register metrics and reset endpoints
	mux.HandleFunc("GET /admin/metrics", apiCfg.MetricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.ResetHandler)
	mux.HandleFunc("POST /api/chirps", api.CreateChirpHandler(dbQueries))
	mux.HandleFunc("POST /api/users", api.CreateUserHandler(dbQueries))
	mux.HandleFunc("GET /api/chirps", api.ListChirpsHandler(dbQueries))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Starting server on port 8080")
	log.Fatal(server.ListenAndServe())
}
