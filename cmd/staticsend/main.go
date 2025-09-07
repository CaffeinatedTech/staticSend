package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"staticsend/pkg/api"
	"staticsend/pkg/auth"
	"staticsend/pkg/database"
	"staticsend/pkg/email"
	"staticsend/pkg/templates"
	"staticsend/pkg/web"
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

	// Get Turnstile configuration for auth pages
	authTurnstilePublicKey := getEnv("STATICSEND_AUTH_TURNSTILE_PUBLIC_KEY", "")
	authTurnstileSecretKey := getEnv("STATICSEND_AUTH_TURNSTILE_SECRET_KEY", "")
	
	// Create template manager and web handlers
	tm := templates.NewTemplateManager()
	webHandler := web.NewWebHandler(database.DB, tm, authTurnstilePublicKey)
	webAuthHandler := web.NewWebAuthHandler(&database.Database{Connection: database.DB}, secretKey, tm, authTurnstilePublicKey, authTurnstileSecretKey)
	settingsHandler := web.NewSettingsHandler(&database.Database{Connection: database.DB}, tm)
	
	// Create email service
	emailService := createEmailService()
	
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
	
	// Form submission endpoint (public)
	r.Post("/api/v1/submit/{formKey}", submissionHandler.SubmitForm)

	// Web pages
	r.Get("/login", webHandler.LoginPage)
	r.Get("/register", webHandler.RegisterPage)

	// Form-based authentication routes
	r.Post("/auth/register", webAuthHandler.RegisterForm)
	r.Post("/auth/login", webAuthHandler.LoginForm)
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

// createEmailService creates and configures the email service
func createEmailService() *email.EmailService {
	// Get SMTP configuration from environment variables
	host := getEnv("STATICSEND_SMTP_HOST", "")
	portStr := getEnv("STATICSEND_SMTP_PORT", "587")
	username := getEnv("STATICSEND_SMTP_USER", "")
	password := getEnv("STATICSEND_SMTP_PASS", "")
	from := getEnv("STATICSEND_SMTP_FROM", username)
	
	// Convert port to integer
	port, err := strconv.Atoi(portStr)
	if err != nil {
		port = 587 // default port
	}
	
	// If no SMTP configuration is provided, create a dummy service that logs emails
	if host == "" || username == "" || password == "" {
		log.Println("WARNING: No SMTP configuration found. Emails will be logged to console instead of being sent.")
		// Create a minimal config for development
		config := email.EmailConfig{
			Host:     "localhost",
			Port:     587,
			Username: "noreply@example.com",
			Password: "",
			From:     "noreply@example.com",
			UseTLS:   false,
		}
		return email.NewEmailService(config, 100, 5, 3)
	}
	
	config := email.EmailConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		UseTLS:   true, // Assume TLS for production
	}
	
	// Test the connection
	if err := email.NewEmailService(config, 100, 5, 3).TestConnection(); err != nil {
		log.Printf("WARNING: SMTP connection test failed: %v", err)
		log.Println("Emails may not be sent successfully. Check your SMTP configuration.")
	}
	
	return email.NewEmailService(config, 100, 5, 3)
}
