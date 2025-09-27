package web

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Run initial migration
	migrationSQL, err := os.ReadFile("../../migrations/001_initial_schema.up.sql")
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Run app settings migration
	migrationSQL, err = os.ReadFile("../../migrations/002_app_settings.up.sql")
	if err != nil {
		t.Fatalf("Failed to read app settings migration file: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("Failed to execute app settings migration: %v", err)
	}

	// Run form schema update migration (adds columns like domain, forward_email, etc.)
	migrationSQL, err = os.ReadFile("../../migrations/003_update_form_schema.up.sql")
	if err != nil {
		t.Fatalf("Failed to read form schema update migration file: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("Failed to execute form schema update migration: %v", err)
	}

	// Run turnstile key fix migration if applicable
	migrationSQL, err = os.ReadFile("../../migrations/004_fix_turnstile_key.up.sql")
	if err != nil {
		t.Fatalf("Failed to read turnstile key fix migration file: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("Failed to execute turnstile key fix migration: %v", err)
	}

    // Run callback URL migration
    migrationSQL, err = os.ReadFile("../../migrations/005_add_callback_url.up.sql")
    if err != nil {
        t.Fatalf("Failed to read callback_url migration file: %v", err)
    }

    if _, err := db.Exec(string(migrationSQL)); err != nil {
        t.Fatalf("Failed to execute callback_url migration: %v", err)
    }

	return db
}

func cleanupTestDB(t *testing.T, db *sql.DB) {
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}
