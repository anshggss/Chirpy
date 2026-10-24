package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/anshggss/goservers/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("SECRET")
	fmt.Printf("dbUrl : %s\n", dbURL)

	// Can connect to any db using this
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Printf("Error while opening connection to databaes: %s", err.Error())
		os.Exit(1)
	}

	// Wraps the given db to perform generated queries on it
	dbQueries := database.New(db)

	router := http.NewServeMux()

	cfg := apiConfig{
		serverRequests: atomic.Int32{},
		db:             dbQueries,
		secret:         secret,
	}

	// Register Handlers to the server

	router.Handle("/app/", cfg.middlewareMetrics(http.StripPrefix("/app/", http.FileServer(http.Dir("./")))))
	router.HandleFunc("GET /api/healthz", health)
	router.HandleFunc("GET /admin/metrics", cfg.metrics)
	router.HandleFunc("POST /admin/reset", cfg.reset)
	router.HandleFunc("POST /api/users", cfg.createUser)
	router.HandleFunc("POST /api/login", cfg.loginUser)
	router.HandleFunc("POST /api/chirps", cfg.createChirps)
	router.HandleFunc("GET /api/chirps/{chirpID}", cfg.getChirp)
	router.HandleFunc("GET /api/chirps", cfg.getChirps)
	router.HandleFunc("POST /api/refresh", cfg.createNewToken)
	router.HandleFunc("POST /api/revoke", cfg.revokeToken)
	router.HandleFunc("PUT /api/users", cfg.updateUser)
	router.HandleFunc("DELETE /api/chirps/{chirpID}", cfg.deleteChirp)
	router.HandleFunc("POST /api/polka/webhooks", cfg.upgradeUser)

	server := http.Server{
		Handler: router,
		Addr:    ":8080",
	}
	fmt.Printf("Serving on %s\n", server.Addr)
	server.ListenAndServe()
}
