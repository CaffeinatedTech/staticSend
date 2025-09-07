package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// Database represents the database connection
type Database struct {
	Connection *sql.DB
}

// DB is the global database connection
var DB *sql.DB

// Init initializes the database connection and runs migrations
func Init(dbPath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	log.Printf("Creating database directory: %s", dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Check if directory is writable
	testFile := filepath.Join(dir, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("database directory is not writable: %w", err)
	}
	os.Remove(testFile)

	log.Printf("Opening database at: %s", dbPath)
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	log.Printf("Database connected: %s", dbPath)

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// runMigrations executes database migrations
func runMigrations() error {
	// Check if users table exists to determine if migrations are needed
	var tableName string
	err := DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)

	if err == sql.ErrNoRows {
		// Tables don't exist, run initial migration
		log.Println("Running initial database migration...")
		migrationSQL, err := os.ReadFile("migrations/001_initial_schema.up.sql")
		if err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		if _, err := DB.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}

		log.Println("Initial migration completed successfully")
	} else if err != nil {
		return fmt.Errorf("failed to check for existing tables: %w", err)
	}

	// Check if app_settings table exists to determine if we need to run the second migration
	var settingsTableName string
	err = DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='app_settings'").Scan(&settingsTableName)

	if err == sql.ErrNoRows {
		// app_settings table doesn't exist, run second migration
		log.Println("Running app settings migration...")
		migrationSQL, err := os.ReadFile("migrations/002_app_settings.up.sql")
		if err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		if _, err := DB.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}

		log.Println("App settings migration completed successfully")
	} else if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check for app_settings table: %w", err)
	}

	// Check if forms table has the new domain column to determine if we need to run the third migration
	var domainColumn string
	err = DB.QueryRow("SELECT name FROM pragma_table_info('forms') WHERE name = 'domain'").Scan(&domainColumn)
	if err == sql.ErrNoRows {
		// forms table doesn't have domain column, run third migration
		log.Println("Running form schema update migration...")
		migrationSQL, err := os.ReadFile("migrations/003_update_form_schema.up.sql")
		if err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		if _, err := DB.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}

		log.Println("Form schema update migration completed successfully")
	} else if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check for forms table columns: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}