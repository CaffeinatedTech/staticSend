package models

import (
	"database/sql"
	"time"
)

// Form represents a contact form configuration
type Form struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Name            string    `json:"name"`
	Domain          string    `json:"domain"`
	TurnstileSecret string    `json:"turnstile_secret"` // Private key for validation
	ForwardEmail    string    `json:"forward_email"`
	FormKey         string    `json:"form_key"`         // Generated unique key
	SubmissionCount int       `json:"submission_count"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CreateForm creates a new form in the database
func CreateForm(db *sql.DB, userID int64, name, domain, turnstileSecret, forwardEmail, formKey string) (*Form, error) {
	result, err := db.Exec(
		"INSERT INTO forms (user_id, name, domain, turnstile_secret, forward_email, form_key) VALUES (?, ?, ?, ?, ?, ?)",
		userID, name, domain, turnstileSecret, forwardEmail, formKey,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return GetFormByID(db, id)
}

// GetFormByID retrieves a form by its ID
func GetFormByID(db *sql.DB, id int64) (*Form, error) {
	var form Form
	err := db.QueryRow(
		"SELECT id, user_id, name, domain, turnstile_secret, forward_email, form_key, created_at, updated_at FROM forms WHERE id = ?",
		id,
	).Scan(&form.ID, &form.UserID, &form.Name, &form.Domain, &form.TurnstileSecret, &form.ForwardEmail, &form.FormKey, &form.CreatedAt, &form.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &form, nil
}

// GetFormsByUserID retrieves all forms for a specific user
func GetFormsByUserID(db *sql.DB, userID int64) ([]Form, error) {
	rows, err := db.Query(
		"SELECT id, user_id, name, domain, turnstile_secret, forward_email, form_key, created_at, updated_at FROM forms WHERE user_id = ? ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forms []Form
	for rows.Next() {
		var form Form
		if err := rows.Scan(&form.ID, &form.UserID, &form.Name, &form.Domain, &form.TurnstileSecret, &form.ForwardEmail, &form.FormKey, &form.CreatedAt, &form.UpdatedAt); err != nil {
			return nil, err
		}
		forms = append(forms, form)
	}

	return forms, nil
}

// FormExists checks if a form with the given name already exists for a user
func FormExists(db *sql.DB, userID int64, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM forms WHERE user_id = ? AND name = ?)",
		userID, name,
	).Scan(&exists)

	return exists, err
}

// GetFormByKey retrieves a form by its form_key
func GetFormByKey(db *sql.DB, formKey string) (*Form, error) {
	var form Form
	err := db.QueryRow(
		"SELECT id, user_id, name, domain, turnstile_secret, forward_email, form_key, created_at, updated_at FROM forms WHERE form_key = ?",
		formKey,
	).Scan(&form.ID, &form.UserID, &form.Name, &form.Domain, &form.TurnstileSecret, &form.ForwardEmail, &form.FormKey, &form.CreatedAt, &form.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &form, nil
}

// UpdateForm updates a form in the database
func UpdateForm(db *sql.DB, formID int64, name, domain, turnstileSecret, forwardEmail string) error {
	_, err := db.Exec(
		"UPDATE forms SET name = ?, domain = ?, turnstile_secret = ?, forward_email = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, domain, turnstileSecret, forwardEmail, formID,
	)
	return err
}