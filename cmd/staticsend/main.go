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
	"staticsend/pkg/config"
	"staticsend/pkg/database"
	"staticsend/pkg/email"
	"staticsend/pkg/templates"
	"staticsend/pkg/web"
	customMiddleware "staticsend/pkg/middleware"
)

func main() {
	// Load configuration from environment variables
	cfg := config.LoadConfig()
	
	// Allow command line overrides
	port := flag.String("port", cfg.Port, "Port to listen on")
	dbPath := flag.String("db", cfg.DatabasePath, "Database file path")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}
	
	// Update config with command line values
	cfg.Port = *port
	cfg.DatabasePath = *dbPath

	// Initialize database
	if err := database.Init(cfg.DatabasePath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Use JWT secret from config
	secretKey := []byte(cfg.JWTSecretKey)

	// Use Turnstile configuration from config
	authTurnstilePublicKey := cfg.TurnstilePublicKey
	authTurnstileSecretKey := cfg.TurnstileSecretKey
	
	// Create template manager and web handlers
	tm := templates.NewTemplateManager()
	webHandler := web.NewWebHandler(database.DB, tm, authTurnstilePublicKey)
	webAuthHandler := web.NewWebAuthHandler(&database.Database{Connection: database.DB}, secretKey, tm, authTurnstilePublicKey, authTurnstileSecretKey)
	settingsHandler := web.NewSettingsHandler(&database.Database{Connection: database.DB}, tm)
	
	// Create email service from config
	emailConfig := email.EmailConfig{
		Host:     cfg.EmailHost,
		Port:     cfg.EmailPort,
		Username: cfg.EmailUsername,
		Password: cfg.EmailPassword,
		From:     cfg.EmailFrom,
		UseTLS:   cfg.EmailUseTLS,
	}
	emailService := email.NewEmailService(emailConfig, 100, 10, 5)
	
	// Create API handlers
	formHandler := api.NewFormHandler(database.DB)
	submissionHandler := api.NewSubmissionHandler(database.DB, emailService)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	// Serve static files
	staticDir := "./static"
	if _, err := os.Stat(staticDir); err == nil {
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	}
	
	// Serve favicon
	r.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/favicon.svg")
	})

	// Public routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	
	// Form submission endpoint (public) with rate limiting
	r.With(customMiddleware.IPRateLimit(time.Minute, 10)).Post("/api/v1/submit/{formKey}", submissionHandler.SubmitForm)

	// Web pages
	r.Get("/login", webHandler.LoginPage)
	r.Get("/register", webHandler.RegisterPage)

	// Form-based authentication routes with rate limiting
	r.With(customMiddleware.IPRateLimit(time.Minute, 5)).Post("/auth/register", webAuthHandler.RegisterForm)
	r.With(customMiddleware.IPRateLimit(time.Minute, 10)).Post("/auth/login", webAuthHandler.LoginForm)
	r.Get("/auth/logout", webAuthHandler.Logout)

	// Protected routes (require authentication)
	r.Group(func(r chi.Router) {
		r.Use(customMiddleware.AuthMiddleware(customMiddleware.AuthConfig{
			SecretKey: secretKey,
			DB:        &database.Database{Connection: database.DB},
			PublicPaths: []string{"/login", "/register", "/health"},
		}))

		r.Get("/", webHandler.Dashboard) // Root route now protected
		r.Get("/dashboard", webHandler.Dashboard)
		r.Get("/settings", settingsHandler.SettingsPage)
		r.Post("/settings/update", settingsHandler.UpdateSettings)
		r.Get("/forms/new", webHandler.CreateFormModal)
		r.Get("/forms/{id}/view", webHandler.ViewFormModal)
		r.Get("/forms/{id}/edit", webHandler.EditFormModal)
		r.Get("/forms/{id}/submissions", webHandler.FormSubmissions)
		
		// Form API routes
		r.Post("/forms", formHandler.CreateForm)
		r.Get("/forms/{id}", formHandler.GetForm)
		r.Put("/forms/{id}", formHandler.UpdateForm)
		r.Delete("/forms/{id}", formHandler.DeleteForm)
		r.Get("/api/forms", formHandler.GetUserForms)
	})

	// Test endpoint for rate limiting
	r.With(customMiddleware.IPRateLimit(time.Second, 2)).Get("/test-rate-limit", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Rate limited endpoint - you should see this only 2 times per second per IP"))
	})

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
