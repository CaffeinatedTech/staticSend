package web

import (
	"testing"

	"staticsend/pkg/templates"
)

func TestWebHandler_NewWebHandler(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create template manager
	tm := &templates.TemplateManager{}

	// Create handler
	handler := NewWebHandler(db, tm, "test-public-key")

	if handler == nil {
		t.Error("NewWebHandler should not return nil")
	}

	if handler.DB == nil {
		t.Error("Handler DB should not be nil")
	}

	if handler.AuthTurnstilePublicKey != "test-public-key" {
		t.Errorf("Expected AuthTurnstilePublicKey 'test-public-key', got '%s'", handler.AuthTurnstilePublicKey)
	}
}

func TestWebHandler_WithoutTurnstile(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create template manager
	tm := &templates.TemplateManager{}

	// Create handler without Turnstile key
	handler := NewWebHandler(db, tm, "")

	if handler.AuthTurnstilePublicKey != "" {
		t.Errorf("Expected empty AuthTurnstilePublicKey, got '%s'", handler.AuthTurnstilePublicKey)
	}
}
