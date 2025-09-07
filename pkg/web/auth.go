package web

import (
	"context"
	"net/http"

	"staticsend/pkg/auth"
	"staticsend/pkg/database"
	"staticsend/pkg/models"
	"staticsend/pkg/templates"
	"staticsend/pkg/turnstile"
)

// WebAuthHandler handles web-based authentication (form submissions)
type WebAuthHandler struct {
	DB                     *database.Database
	SecretKey              []byte
	Templates              *templates.TemplateManager
	AuthTurnstilePublicKey string
	AuthTurnstileSecretKey string
}

// NewWebAuthHandler creates a new web auth handler
func NewWebAuthHandler(db *database.Database, secretKey []byte, tm *templates.TemplateManager, authTurnstilePublicKey, authTurnstileSecretKey string) *WebAuthHandler {
	return &WebAuthHandler{
		DB:                     db,
		SecretKey:              secretKey,
		Templates:              tm,
		AuthTurnstilePublicKey: authTurnstilePublicKey,
		AuthTurnstileSecretKey: authTurnstileSecretKey,
	}
}

// RegisterForm handles form-based user registration
func (h *WebAuthHandler) RegisterForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderRegisterPage(w, "Invalid form data")
		return
	}

	// Check if registration is enabled
	enabled, err := models.IsRegistrationEnabled(h.DB.Connection)
	if err != nil {
		h.renderRegisterPage(w, "Internal server error")
		return
	}
	if !enabled {
		h.renderRegisterPage(w, "Registration is currently disabled")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate input
	if email == "" || password == "" {
		h.renderRegisterPage(w, "Email and password are required")
		return
	}

	// Validate Turnstile token if configured
	if h.AuthTurnstileSecretKey != "" {
		turnstileToken := r.FormValue("cf-turnstile-response")
		if turnstileToken == "" {
			h.renderRegisterPage(w, "Bot protection verification required")
			return
		}

		validator := turnstile.NewValidator(h.AuthTurnstileSecretKey)
		ctx := context.Background()
		response, err := validator.Verify(ctx, turnstileToken, r.RemoteAddr)
		if err != nil {
			h.renderRegisterPage(w, "Bot protection verification failed")
			return
		}

		if !response.IsValid() {
			h.renderRegisterPage(w, "Bot protection verification failed")
			return
		}
	}

	// Check if user already exists
	exists, err := models.UserExists(h.DB.Connection, email)
	if err != nil {
		h.renderRegisterPage(w, "Internal server error")
		return
	}
	if exists {
		h.renderRegisterPage(w, "User already exists")
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		h.renderRegisterPage(w, "Failed to process password")
		return
	}

	// Create user
	user, err := models.CreateUser(h.DB.Connection, email, passwordHash)
	if err != nil {
		h.renderRegisterPage(w, "Failed to create user")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user, h.SecretKey)
	if err != nil {
		h.renderRegisterPage(w, "Failed to generate token")
		return
	}

	// Set token as cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})

	// Use HX-Redirect for HTMX to properly handle the redirect
	w.Header().Set("HX-Redirect", "/dashboard")
}

// LoginForm handles form-based user login
func (h *WebAuthHandler) LoginForm(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderLoginPage(w, "Invalid form data")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	// Validate input
	if email == "" || password == "" {
		h.renderLoginPage(w, "Email and password are required")
		return
	}

	// Validate Turnstile token if configured
	if h.AuthTurnstileSecretKey != "" {
		turnstileToken := r.FormValue("cf-turnstile-response")
		if turnstileToken == "" {
			h.renderLoginPage(w, "Bot protection verification required")
			return
		}

		validator := turnstile.NewValidator(h.AuthTurnstileSecretKey)
		ctx := context.Background()
		response, err := validator.Verify(ctx, turnstileToken, r.RemoteAddr)
		if err != nil {
			h.renderLoginPage(w, "Bot protection verification failed")
			return
		}

		if !response.IsValid() {
			h.renderLoginPage(w, "Bot protection verification failed")
			return
		}
	}

	// Get user by email
	user, err := models.GetUserByEmail(h.DB.Connection, email)
	if err != nil {
		h.renderLoginPage(w, "Internal server error")
		return
	}
	if user == nil {
		h.renderLoginPage(w, "Invalid email or password")
		return
	}

	// Check password
	if err := auth.CheckPassword(password, user.PasswordHash); err != nil {
		h.renderLoginPage(w, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user, h.SecretKey)
	if err != nil {
		h.renderLoginPage(w, "Failed to generate token")
		return
	}

	// Set token as cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
	})

	// Use HX-Redirect for HTMX to properly handle the redirect
	w.Header().Set("HX-Redirect", "/dashboard")
}

// renderRegisterPage renders the registration page with an optional error
func (h *WebAuthHandler) renderRegisterPage(w http.ResponseWriter, errorMsg string) {
	data := templates.TemplateData{
		Title:                  "Register - staticSend",
		Error:                  errorMsg,
		ShowHeader:             false,
		AuthTurnstilePublicKey: h.AuthTurnstilePublicKey,
	}
	
	h.Templates.Render(w, "auth/register.html", data)
}

// renderLoginPage renders the login page with an optional error
func (h *WebAuthHandler) renderLoginPage(w http.ResponseWriter, errorMsg string) {
	data := templates.TemplateData{
		Title:                  "Login - staticSend",
		Error:                  errorMsg,
		ShowHeader:             false,
		AuthTurnstilePublicKey: h.AuthTurnstilePublicKey,
	}
	
	h.Templates.Render(w, "auth/login.html", data)
}

// Logout handles user logout
func (h *WebAuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the auth cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1, // Immediately expire the cookie
	})

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusFound)
}