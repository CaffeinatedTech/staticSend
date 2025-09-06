package models

import (
	"database/sql"
	"testing"

	"staticsend/pkg/utils"
)

// CreateTestForm is a helper function for tests that creates a form with auto-generated key
func CreateTestForm(t *testing.T, db *sql.DB, userID int64, name, domain, turnstileSecret, forwardEmail string) *Form {
	formKey, err := utils.GenerateFormKey()
	if err != nil {
		t.Fatalf("Failed to generate form key: %v", err)
	}
	
	form, err := CreateForm(db, userID, name, domain, turnstileSecret, forwardEmail, formKey)
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}
	return form
}