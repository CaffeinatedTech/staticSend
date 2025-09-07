package web

import (
	"testing"

	"staticsend/pkg/database"
	"staticsend/pkg/templates"
)

func TestWebAuthHandler_NewWebAuthHandler(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create template manager
	tm := &templates.TemplateManager{}

	// Create handler
	handler := NewWebAuthHandler(&database.Database{Connection: db}, []byte("test-secret"), tm, "", "")

	if handler == nil {
		t.Error("NewWebAuthHandler should not return nil")
	}

	if handler.DB == nil {
		t.Error("Handler DB should not be nil")
	}

	if handler.SecretKey == nil {
		t.Error("Handler SecretKey should not be nil")
	}
}

func TestWebAuthHandler_WithTurnstileKeys(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create template manager
	tm := &templates.TemplateManager{}

	// Create handler with Turnstile keys
	handler := NewWebAuthHandler(&database.Database{Connection: db}, []byte("test-secret"), tm, "test-public-key", "test-secret-key")

	if handler.AuthTurnstilePublicKey != "test-public-key" {
		t.Errorf("Expected AuthTurnstilePublicKey 'test-public-key', got '%s'", handler.AuthTurnstilePublicKey)
	}

	if handler.AuthTurnstileSecretKey != "test-secret-key" {
		t.Errorf("Expected AuthTurnstileSecretKey 'test-secret-key', got '%s'", handler.AuthTurnstileSecretKey)
	}
}
