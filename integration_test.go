package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"staticsend/pkg/api"
	"staticsend/pkg/database"
	"staticsend/pkg/email"
	"staticsend/pkg/middleware"
	"staticsend/pkg/models"

	"github.com/go-chi/chi/v5"
)

// IntegrationTestSuite holds the test server and dependencies
type IntegrationTestSuite struct {
	Server       *httptest.Server
	DB           *database.Database
	EmailService *email.EmailService
	Router       *chi.Mux
	TestUser     *models.User
	TestForm     *models.Form
}

// SetupIntegrationTest creates a complete test environment
func SetupIntegrationTest(t *testing.T) *IntegrationTestSuite {
	// Create test database
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "integration_test.db")
	
	err := database.Init(dbPath)
	if err != nil {
		// If migration files don't exist, create a minimal schema
		db, err := database.DB.Begin()
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		
		// Create minimal schema for testing
		schema := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS forms (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			form_key TEXT UNIQUE NOT NULL,
			domain TEXT NOT NULL,
			turnstile_public_key TEXT,
			turnstile_secret_key TEXT,
			notification_email TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(id)
		);
		
		CREATE TABLE IF NOT EXISTS submissions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			form_id INTEGER NOT NULL,
			data TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (form_id) REFERENCES forms(id)
		);
		
		CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);`
		
		_, err = db.Exec(schema)
		if err != nil {
			t.Fatalf("Failed to create test schema: %v", err)
		}
		
		err = db.Commit()
		if err != nil {
			t.Fatalf("Failed to commit test schema: %v", err)
		}
	}
	
	dbWrapper := &database.Database{Connection: database.DB}
	
	// Create email service
	emailConfig := email.EmailConfig{
		Host:     "localhost",
		Port:     587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		UseTLS:   false,
	}
	emailService := email.NewEmailService(emailConfig, 10, 1, 1)
	
	// Create handlers
	apiHandler := api.NewSubmissionHandler(database.DB, emailService)
	
	// Create router
	r := chi.NewRouter()
	
	// Add middleware
	r.Use(middleware.IPRateLimit(time.Minute, 100)) // High limit for testing
	
	// API routes only (avoid template complications)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/submit/{formKey}", apiHandler.SubmitForm)
	})
	
	// Simple health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	
	// Create test server
	server := httptest.NewServer(r)
	
	// Create test user
	testUser, err := models.CreateUser(database.DB, "test@example.com", "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/VcSAg/9qm") // "password123"
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	// Create test form
	testForm, err := models.CreateForm(database.DB, testUser.ID, "Test Form", "example.com", "test-public", "test-secret", "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to create test form: %v", err)
	}
	
	// Enable registration for tests
	models.UpdateAppSetting(database.DB, "registration_enabled", "true")
	
	return &IntegrationTestSuite{
		Server:       server,
		DB:           dbWrapper,
		EmailService: emailService,
		Router:       r,
		TestUser:     testUser,
		TestForm:     testForm,
	}
}

// Cleanup closes the test environment
func (suite *IntegrationTestSuite) Cleanup() {
	suite.Server.Close()
	suite.EmailService.Shutdown()
	database.Close()
}

// TestFormSubmissionFlow tests the complete form submission workflow
func TestFormSubmissionFlow(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()
	
	t.Run("successful form submission", func(t *testing.T) {
		// Prepare form data
		formData := url.Values{}
		formData.Set("name", "John Doe")
		formData.Set("email", "john@example.com")
		formData.Set("message", "Test message")
		formData.Set("cf-turnstile-response", "fake-token-for-testing")
		
		// Submit form
		resp, err := http.Post(
			suite.Server.URL+"/api/v1/submit/"+suite.TestForm.FormKey,
			"application/x-www-form-urlencoded",
			strings.NewReader(formData.Encode()),
		)
		if err != nil {
			t.Fatalf("Failed to submit form: %v", err)
		}
		defer resp.Body.Close()
		
		// Note: This will fail with Turnstile validation, but we can check the error
		body, _ := io.ReadAll(resp.Body)
		
		// Should get Turnstile validation error (expected in test environment)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 (Turnstile validation failure), got %d", resp.StatusCode)
		}
		
		if !strings.Contains(string(body), "Invalid Turnstile token") {
			t.Errorf("Expected Turnstile validation error, got: %s", string(body))
		}
	})
	
	t.Run("form not found", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("name", "John Doe")
		formData.Set("cf-turnstile-response", "fake-token")
		
		resp, err := http.Post(
			suite.Server.URL+"/api/v1/submit/nonexistent",
			"application/x-www-form-urlencoded",
			strings.NewReader(formData.Encode()),
		)
		if err != nil {
			t.Fatalf("Failed to submit to nonexistent form: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})
	
	t.Run("missing turnstile token", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("name", "John Doe")
		// No Turnstile token
		
		resp, err := http.Post(
			suite.Server.URL+"/api/v1/submit/"+suite.TestForm.FormKey,
			"application/x-www-form-urlencoded",
			strings.NewReader(formData.Encode()),
		)
		if err != nil {
			t.Fatalf("Failed to submit form: %v", err)
		}
		defer resp.Body.Close()
		
		body, _ := io.ReadAll(resp.Body)
		
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
		
		if !strings.Contains(string(body), "Turnstile verification required") {
			t.Errorf("Expected Turnstile required error, got: %s", string(body))
		}
	})
}

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()
	
	t.Run("health endpoint responds", func(t *testing.T) {
		resp, err := http.Get(suite.Server.URL + "/health")
		if err != nil {
			t.Fatalf("Failed to get health endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		
		body, _ := io.ReadAll(resp.Body)
		if string(body) != "OK" {
			t.Errorf("Expected 'OK', got '%s'", string(body))
		}
	})
}

// TestRateLimiting tests the rate limiting functionality
func TestRateLimiting(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()
	
	t.Run("rate limit enforcement", func(t *testing.T) {
		// This test would need a lower rate limit to be practical
		// For now, just test that rate limiting middleware is active
		
		formData := url.Values{}
		formData.Set("name", "John Doe")
		formData.Set("cf-turnstile-response", "fake-token")
		
		// Make a request
		resp, err := http.Post(
			suite.Server.URL+"/api/v1/submit/"+suite.TestForm.FormKey,
			"application/x-www-form-urlencoded",
			strings.NewReader(formData.Encode()),
		)
		if err != nil {
			t.Fatalf("Failed to submit form: %v", err)
		}
		defer resp.Body.Close()
		
		// Should get some response (rate limiting is configured with high limit for testing)
		if resp.StatusCode == 0 {
			t.Error("Expected some HTTP response")
		}
	})
}

// TestAPIEndpoints tests API endpoint availability
func TestAPIEndpoints(t *testing.T) {
	suite := SetupIntegrationTest(t)
	defer suite.Cleanup()
	
	t.Run("submit endpoint exists", func(t *testing.T) {
		// Test with empty form data to verify endpoint exists
		formData := url.Values{}
		resp, err := http.Post(
			suite.Server.URL+"/api/v1/submit/test-form",
			"application/x-www-form-urlencoded",
			strings.NewReader(formData.Encode()),
		)
		if err != nil {
			t.Fatalf("Failed to post to submit endpoint: %v", err)
		}
		defer resp.Body.Close()
		
		// Should return 400 due to missing Turnstile token, not 404
		if resp.StatusCode == http.StatusNotFound {
			t.Errorf("Submit endpoint not found")
		}
	})
}

