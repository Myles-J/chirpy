package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Myles-J/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
}

func (cfg *apiConfig) handlerMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>
	<body>
	<h1>Welcome, Chirpy Admin</h1>
	<p>Chirpy has been visited %d times!</p>
	</body>
	</html>
	`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
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
	w.Write([]byte("Hits reset to 0"))
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	badWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}
	type parameters struct {
		Body string `json:"body"`
	}

	type response struct {
		Valid       bool   `json:"valid"`
		CleanedBody string `json:"cleaned_body,omitempty"`
		Error       string `json:"error,omitempty"`
	}

	w.Header().Set("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)

	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp := response{
			Valid: false,
			Error: "Bad Request",
		}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	if len(params.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		resp := response{
			Valid: false,
			Error: "Chirp is too long",
		}
		jsonResp, _ := json.Marshal(resp)
		w.Write(jsonResp)
		return
	}

	words := strings.Split(params.Body, " ")
	cleanedWords := make([]string, len(words))

	for i, word := range words {
		if _, ok := badWords[strings.ToLower(word)]; ok {
			cleanedWords[i] = "****"
		} else {
			cleanedWords[i] = word
		}
	}

	cleanedBody := strings.Join(cleanedWords, " ")

	w.WriteHeader(http.StatusOK)
	resp := response{
		Valid:       true,
		CleanedBody: cleanedBody,
	}
	jsonResp, _ := json.Marshal(resp)
	w.Write(jsonResp)
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

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

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
	}

	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.handlerMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})))

	mux.HandleFunc("GET /api/healthz", handleHealthCheck)

	// Serve static files from the assets directory
	fs := http.FileServer(http.Dir("assets"))
	mux.Handle("/app/assets/", apiCfg.handlerMetrics(http.StripPrefix("/app/assets/", fs)))

	// Register metrics and reset endpoints
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHandler)

	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	mux.HandleFunc("POST /api/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			Email string `json:"email"`
		}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dbUser, err := apiCfg.db.CreateUser(context.Background(), params.Email)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user := User{
			ID:        dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email:     dbUser.Email,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("Starting server on port 8080")
	log.Fatal(server.ListenAndServe())
}
