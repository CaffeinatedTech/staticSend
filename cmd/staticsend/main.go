package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"staticsend/pkg/api"
	"staticsend/pkg/auth"
	"staticsend/pkg/database"
	customMiddleware "staticsend/pkg/middleware"
)

func main() {
	port := flag.String("port", getEnv("STATICSEND_PORT", "8080"), "Port to listen on")
	dbPath := flag.String("db", getEnv("STATICSEND_DB_PATH", "./data/staticsend.db"), "Database file path")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	// Initialize database
	if err := database.Init(*dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Generate or load secret key for JWT
	secretKey := getSecretKey()

	// Create auth handler
	authHandler := &api.AuthHandler{
		DB:        &database.Database{Connection: database.DB},
		SecretKey: secretKey,
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Public routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("staticSend API Server"))
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// Authentication routes
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Get("/health", authHandler.HealthCheck)
	})

	// Test endpoint for rate limiting
	r.With(customMiddleware.IPRateLimit(time.Second, 2)).Get("/test-rate-limit", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Rate limited endpoint - you should see this only 2 times per second per IP"))
	})

	log.Printf("Starting server on :%s", *port)
	if err := http.ListenAndServe(":"+*port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getSecretKey retrieves or generates the JWT secret key
func getSecretKey() []byte {
	// Try to get from environment variable
	if envKey := os.Getenv("STATICSEND_JWT_SECRET"); envKey != "" {
		return []byte(envKey)
	}

	// For development, generate a new key each time
	// In production, this should be a persistent secret
	key, err := auth.GenerateSecretKey()
	if err != nil {
		log.Fatalf("Failed to generate secret key: %v", err)
	}
	
	log.Println("WARNING: Using auto-generated JWT secret key. For production, set STATICSEND_JWT_SECRET environment variable.")
	return key
}
