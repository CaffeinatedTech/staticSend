package models

import (
	"database/sql"
	"time"
)

// AppSetting represents an application-wide setting
type AppSetting struct {
	ID          int64     `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// GetAppSetting retrieves an application setting by key
func GetAppSetting(db *sql.DB, key string) (*AppSetting, error) {
	var setting AppSetting
	err := db.QueryRow(
		"SELECT id, key, value, description, created_at, updated_at FROM app_settings WHERE key = ?",
		key,
	).Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.CreatedAt, &setting.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

// GetAppSettingValue retrieves just the value of an application setting by key
func GetAppSettingValue(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow(
		"SELECT value FROM app_settings WHERE key = ?",
		key,
	).Scan(&value)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	return value, nil
}

// GetAppSettingBool retrieves a boolean application setting by key
func GetAppSettingBool(db *sql.DB, key string) (bool, error) {
	value, err := GetAppSettingValue(db, key)
	if err != nil {
		return false, err
	}
	
	return value == "true", nil
}

// UpdateAppSetting updates an application setting
func UpdateAppSetting(db *sql.DB, key, value string) error {
	_, err := db.Exec(
		"UPDATE app_settings SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE key = ?",
		value, key,
	)
	return err
}

// GetAllAppSettings retrieves all application settings
func GetAllAppSettings(db *sql.DB) ([]AppSetting, error) {
	rows, err := db.Query(
		"SELECT id, key, value, description, created_at, updated_at FROM app_settings ORDER BY key",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []AppSetting
	for rows.Next() {
		var setting AppSetting
		if err := rows.Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.CreatedAt, &setting.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, nil
}

// IsRegistrationEnabled checks if user registration is enabled
func IsRegistrationEnabled(db *sql.DB) (bool, error) {
	return GetAppSettingBool(db, "registration_enabled")
}