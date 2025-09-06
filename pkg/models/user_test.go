package models

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
		t.Fatalf("Failed to read migration file: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	// Run form schema update migration
	migrationSQL, err = os.ReadFile("../../migrations/003_update_form_schema.up.sql")
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}

	return db
}

func TestCreateUser(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test creating a new user
	user, err := CreateUser(db, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
	}

	if user.PasswordHash != "hashed_password" {
		t.Errorf("Expected password hash 'hashed_password', got '%s'", user.PasswordHash)
	}

	// Test duplicate email
	_, err = CreateUser(db, "test@example.com", "another_hash")
	if err == nil {
		t.Error("Expected error when creating user with duplicate email")
	}
}

func TestGetUserByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	createdUser, err := CreateUser(db, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test retrieving the user
	retrievedUser, err := GetUserByID(db, createdUser.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}

	if retrievedUser == nil {
		t.Fatal("Expected to retrieve user, got nil")
	}

	if retrievedUser.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", retrievedUser.Email)
	}

	// Test non-existent user
	nonExistentUser, err := GetUserByID(db, 999)
	if err != nil {
		t.Fatalf("Unexpected error getting non-existent user: %v", err)
	}

	if nonExistentUser != nil {
		t.Error("Expected nil for non-existent user")
	}
}

func TestGetUserByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	createdUser, err := CreateUser(db, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test retrieving the user by email
	retrievedUser, err := GetUserByEmail(db, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	if retrievedUser == nil {
		t.Fatal("Expected to retrieve user, got nil")
	}

	if retrievedUser.ID != createdUser.ID {
		t.Errorf("Expected ID %d, got %d", createdUser.ID, retrievedUser.ID)
	}

	// Test non-existent email
	nonExistentUser, err := GetUserByEmail(db, "nonexistent@example.com")
	if err != nil {
		t.Fatalf("Unexpected error getting non-existent user: %v", err)
	}

	if nonExistentUser != nil {
		t.Error("Expected nil for non-existent email")
	}
}

func TestUserExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	_, err := CreateUser(db, "test@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test existing user
	exists, err := UserExists(db, "test@example.com")
	if err != nil {
		t.Fatalf("Failed to check user existence: %v", err)
	}

	if !exists {
		t.Error("Expected user to exist")
	}

	// Test non-existent user
	exists, err = UserExists(db, "nonexistent@example.com")
	if err != nil {
		t.Fatalf("Failed to check non-existent user: %v", err)
	}

	if exists {
		t.Error("Expected user to not exist")
	}
}