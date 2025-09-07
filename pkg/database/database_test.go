package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// SetupTestDB creates a test database for testing
func SetupTestDB(t *testing.T) *sql.DB {
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

	return db
}

// CleanupTestDB closes the test database
func CleanupTestDB(t *testing.T, db *sql.DB) {
	if err := db.Close(); err != nil {
		t.Errorf("Failed to close test database: %v", err)
	}
}

func TestDatabase_Struct(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupTestDB(t, db)

	database := &Database{Connection: db}
	
	if database.Connection == nil {
		t.Error("Expected Connection to be set")
	}
}

func TestClose_NilDB(t *testing.T) {
	// Save original DB
	originalDB := DB
	defer func() { DB = originalDB }()

	// Set DB to nil
	DB = nil

	err := Close()
	if err != nil {
		t.Errorf("Close should not return error when DB is nil, got: %v", err)
	}
}
