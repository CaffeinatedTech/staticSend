package models

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupSettingsTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Run migrations
	migrationSQL, err := os.ReadFile("../../migrations/001_initial_schema.up.sql")
	if err != nil {
		t.Fatalf("Failed to read initial migration: %v", err)
	}
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("Failed to execute initial migration: %v", err)
	}

	migrationSQL, err = os.ReadFile("../../migrations/002_app_settings.up.sql")
	if err != nil {
		t.Fatalf("Failed to read settings migration: %v", err)
	}
	if _, err := db.Exec(string(migrationSQL)); err != nil {
		t.Fatalf("Failed to execute settings migration: %v", err)
	}

	return db
}

func TestAppSettings(t *testing.T) {
	db := setupSettingsTestDB(t)
	defer db.Close()

	t.Run("GetAppSetting", func(t *testing.T) {
		setting, err := GetAppSetting(db, "registration_enabled")
		if err != nil {
			t.Fatalf("Failed to get setting: %v", err)
		}

		if setting.Key != "registration_enabled" {
			t.Errorf("Expected key 'registration_enabled', got '%s'", setting.Key)
		}

		if setting.Value != "true" {
			t.Errorf("Expected value 'true', got '%s'", setting.Value)
		}
	})

	t.Run("GetAppSettingValue", func(t *testing.T) {
		value, err := GetAppSettingValue(db, "registration_enabled")
		if err != nil {
			t.Fatalf("Failed to get setting value: %v", err)
		}

		if value != "true" {
			t.Errorf("Expected value 'true', got '%s'", value)
		}
	})

	t.Run("GetAppSettingBool", func(t *testing.T) {
		enabled, err := GetAppSettingBool(db, "registration_enabled")
		if err != nil {
			t.Fatalf("Failed to get setting bool: %v", err)
		}

		if !enabled {
			t.Error("Expected registration to be enabled")
		}
	})

	t.Run("UpdateAppSetting", func(t *testing.T) {
		err := UpdateAppSetting(db, "registration_enabled", "false")
		if err != nil {
			t.Fatalf("Failed to update setting: %v", err)
		}

		enabled, err := GetAppSettingBool(db, "registration_enabled")
		if err != nil {
			t.Fatalf("Failed to get updated setting: %v", err)
		}

		if enabled {
			t.Error("Expected registration to be disabled after update")
		}
	})

	t.Run("IsRegistrationEnabled", func(t *testing.T) {
		// Reset to enabled
		err := UpdateAppSetting(db, "registration_enabled", "true")
		if err != nil {
			t.Fatalf("Failed to update setting: %v", err)
		}

		enabled, err := IsRegistrationEnabled(db)
		if err != nil {
			t.Fatalf("Failed to check registration status: %v", err)
		}

		if !enabled {
			t.Error("Expected registration to be enabled")
		}

		// Test disabled
		err = UpdateAppSetting(db, "registration_enabled", "false")
		if err != nil {
			t.Fatalf("Failed to update setting: %v", err)
		}

		enabled, err = IsRegistrationEnabled(db)
		if err != nil {
			t.Fatalf("Failed to check registration status: %v", err)
		}

		if enabled {
			t.Error("Expected registration to be disabled")
		}
	})

	t.Run("GetAllAppSettings", func(t *testing.T) {
		settings, err := GetAllAppSettings(db)
		if err != nil {
			t.Fatalf("Failed to get all settings: %v", err)
		}

		if len(settings) != 3 {
			t.Errorf("Expected 3 settings, got %d", len(settings))
		}

		// Check that we have the expected keys
		expectedKeys := map[string]bool{
			"registration_enabled": false,
			"site_title":          false,
			"site_description":    false,
		}

		for _, setting := range settings {
			expectedKeys[setting.Key] = true
		}

		for key, found := range expectedKeys {
			if !found {
				t.Errorf("Expected setting key '%s' not found", key)
			}
		}
	})
}