package turnstile

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator("test-secret-key")

	if validator.secretKey != "test-secret-key" {
		t.Errorf("Expected secret key 'test-secret-key', got '%s'", validator.secretKey)
	}

	if validator.verifyURL != DefaultVerifyURL {
		t.Errorf("Expected verify URL '%s', got '%s'", DefaultVerifyURL, validator.verifyURL)
	}

	if validator.httpClient.Timeout != DefaultTimeout {
		t.Errorf("Expected timeout %v, got %v", DefaultTimeout, validator.httpClient.Timeout)
	}
}

func TestValidator_WithVerifyURL(t *testing.T) {
	validator := NewValidator("test-secret-key")
	customURL := "https://custom-verify.example.com"

	validator = validator.WithVerifyURL(customURL)

	if validator.verifyURL != customURL {
		t.Errorf("Expected custom verify URL '%s', got '%s'", customURL, validator.verifyURL)
	}
}

func TestValidator_Verify_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request content type
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("Expected Content-Type application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Verify form data
		if r.Form.Get("secret") != "test-secret" {
			http.Error(w, "Invalid secret", http.StatusBadRequest)
			return
		}

		if r.Form.Get("response") != "valid-token" {
			http.Error(w, "Invalid token", http.StatusBadRequest)
			return
		}

		if r.Form.Get("remoteip") != "192.168.1.1" {
			http.Error(w, "Invalid remote IP", http.StatusBadRequest)
			return
		}

		// Return success response
		response := VerificationResponse{
			Success:     true,
			ChallengeTS: "2023-01-01T00:00:00Z",
			Hostname:    "example.com",
			ErrorCodes:  []string{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create validator with test server URL
	validator := NewValidator("test-secret").WithVerifyURL(server.URL)

	// Test verification
	ctx := context.Background()
	response, err := validator.Verify(ctx, "valid-token", "192.168.1.1")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !response.Success {
		t.Error("Expected verification to succeed")
	}

	if response.Hostname != "example.com" {
		t.Errorf("Expected hostname 'example.com', got '%s'", response.Hostname)
	}

	if len(response.ErrorCodes) != 0 {
		t.Errorf("Expected no error codes, got %v", response.ErrorCodes)
	}
}

func TestValidator_Verify_Failure(t *testing.T) {
	// Create test server that returns failure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := VerificationResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-input-response"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	validator := NewValidator("test-secret").WithVerifyURL(server.URL)

	ctx := context.Background()
	response, err := validator.Verify(ctx, "invalid-token", "")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Success {
		t.Error("Expected verification to fail")
	}

	if !response.HasError("invalid-input-response") {
		t.Error("Expected invalid-input-response error code")
	}
}

func TestValidator_Verify_EmptyToken(t *testing.T) {
	validator := NewValidator("test-secret")

	ctx := context.Background()
	response, err := validator.Verify(ctx, "", "192.168.1.1")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Success {
		t.Error("Expected verification to fail with empty token")
	}

	if !response.HasError("missing-input-response") {
		t.Error("Expected missing-input-response error code")
	}
}

func TestValidator_Verify_NetworkError(t *testing.T) {
	// Create validator with invalid URL to simulate network error
	validator := NewValidator("test-secret").WithVerifyURL("https://invalid-domain-that-does-not-exist-12345.com")

	ctx := context.Background()
	_, err := validator.Verify(ctx, "test-token", "")

	if err == nil {
		t.Error("Expected network error")
	}
}

func TestValidator_Verify_InvalidJSONResponse(t *testing.T) {
	// Create test server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	validator := NewValidator("test-secret").WithVerifyURL(server.URL)

	ctx := context.Background()
	_, err := validator.Verify(ctx, "test-token", "")

	if err == nil {
		t.Error("Expected JSON parsing error")
	}
}

func TestVerificationResponse_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		response *VerificationResponse
		expected bool
	}{
		{
			name: "successful verification",
			response: &VerificationResponse{
				Success: true,
			},
			expected: true,
		},
		{
			name: "failed verification",
			response: &VerificationResponse{
				Success: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.response.IsValid()
			if result != tt.expected {
				t.Errorf("Expected IsValid() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestVerificationResponse_HasError(t *testing.T) {
	response := &VerificationResponse{
		ErrorCodes: []string{"invalid-input-response", "timeout-or-duplicate"},
	}

	if !response.HasError("invalid-input-response") {
		t.Error("Expected to find invalid-input-response error")
	}

	if !response.HasError("timeout-or-duplicate") {
		t.Error("Expected to find timeout-or-duplicate error")
	}

	if response.HasError("nonexistent-error") {
		t.Error("Should not find nonexistent error")
	}
}

func TestValidator_Verify_Timeout(t *testing.T) {
	// Create test server that delays response to trigger timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Longer than test client timeout

		response := VerificationResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create validator with very short timeout
	validator := NewValidator("test-secret")
	validator.httpClient.Timeout = 10 * time.Millisecond
	validator = validator.WithVerifyURL(server.URL)

	ctx := context.Background()
	_, err := validator.Verify(ctx, "test-token", "")

	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestValidator_Verify_ContextCancellation(t *testing.T) {
	// Create test server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)

		response := VerificationResponse{
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	validator := NewValidator("test-secret").WithVerifyURL(server.URL)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := validator.Verify(ctx, "test-token", "")

	if err == nil {
		t.Error("Expected context cancellation error")
	}
}
