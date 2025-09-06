package models

import (
	"encoding/json"
	"testing"
)

func TestCreateSubmissionEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form, err := CreateForm(db, user.ID, "contact", "Contact Form", "Test form", "")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	submissionData := map[string]interface{}{"test": "data"}
	dataBytes, _ := json.Marshal(submissionData)
	submission, err := CreateSubmission(db, form.ID, "192.168.1.1", "Test Browser", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	// Test creating a new email record
	email, err := CreateSubmissionEmail(db, submission.ID, "sent", "")
	if err != nil {
		t.Fatalf("Failed to create submission email: %v", err)
	}

	if email.SubmissionID != submission.ID {
		t.Errorf("Expected submission ID %d, got %d", submission.ID, email.SubmissionID)
	}

	if email.Status != "sent" {
		t.Errorf("Expected status 'sent', got '%s'", email.Status)
	}

	if email.ErrorMessage != "" {
		t.Errorf("Expected empty error message, got '%s'", email.ErrorMessage)
	}

	// Test creating email record with error
	emailWithError, err := CreateSubmissionEmail(db, submission.ID, "failed", "SMTP error")
	if err != nil {
		t.Fatalf("Failed to create submission email with error: %v", err)
	}

	if emailWithError.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", emailWithError.Status)
	}

	if emailWithError.ErrorMessage != "SMTP error" {
		t.Errorf("Expected error message 'SMTP error', got '%s'", emailWithError.ErrorMessage)
	}
}

func TestGetSubmissionEmailByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form, err := CreateForm(db, user.ID, "contact", "Contact Form", "Test form", "")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	submissionData := map[string]interface{}{"test": "data"}
	dataBytes, _ := json.Marshal(submissionData)
	submission, err := CreateSubmission(db, form.ID, "192.168.1.1", "Test Browser", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	createdEmail, err := CreateSubmissionEmail(db, submission.ID, "sent", "")
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	// Test retrieving the email record
	retrievedEmail, err := GetSubmissionEmailByID(db, createdEmail.ID)
	if err != nil {
		t.Fatalf("Failed to get email by ID: %v", err)
	}

	if retrievedEmail == nil {
		t.Fatal("Expected to retrieve email, got nil")
	}

	if retrievedEmail.ID != createdEmail.ID {
		t.Errorf("Expected ID %d, got %d", createdEmail.ID, retrievedEmail.ID)
	}

	// Test non-existent email
	nonExistentEmail, err := GetSubmissionEmailByID(db, 999)
	if err != nil {
		t.Fatalf("Unexpected error getting non-existent email: %v", err)
	}

	if nonExistentEmail != nil {
		t.Error("Expected nil for non-existent email")
	}
}

func TestGetSubmissionEmailBySubmissionID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form, err := CreateForm(db, user.ID, "contact", "Contact Form", "Test form", "")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	submissionData := map[string]interface{}{"test": "data"}
	dataBytes, _ := json.Marshal(submissionData)
	submission, err := CreateSubmission(db, form.ID, "192.168.1.1", "Test Browser", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	createdEmail, err := CreateSubmissionEmail(db, submission.ID, "sent", "")
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	// Test retrieving the email record by submission ID
	retrievedEmail, err := GetSubmissionEmailBySubmissionID(db, submission.ID)
	if err != nil {
		t.Fatalf("Failed to get email by submission ID: %v", err)
	}

	if retrievedEmail == nil {
		t.Fatal("Expected to retrieve email, got nil")
	}

	if retrievedEmail.ID != createdEmail.ID {
		t.Errorf("Expected ID %d, got %d", createdEmail.ID, retrievedEmail.ID)
	}

	// Test non-existent submission ID
	nonExistentEmail, err := GetSubmissionEmailBySubmissionID(db, 999)
	if err != nil {
		t.Fatalf("Unexpected error getting email for non-existent submission: %v", err)
	}

	if nonExistentEmail != nil {
		t.Error("Expected nil for non-existent submission ID")
	}
}

func TestUpdateSubmissionEmailStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form, err := CreateForm(db, user.ID, "contact", "Contact Form", "Test form", "")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	submissionData := map[string]interface{}{"test": "data"}
	dataBytes, _ := json.Marshal(submissionData)
	submission, err := CreateSubmission(db, form.ID, "192.168.1.1", "Test Browser", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	createdEmail, err := CreateSubmissionEmail(db, submission.ID, "sent", "")
	if err != nil {
		t.Fatalf("Failed to create email: %v", err)
	}

	// Test updating email status
	err = UpdateSubmissionEmailStatus(db, createdEmail.ID, "failed", "SMTP connection failed")
	if err != nil {
		t.Fatalf("Failed to update email status: %v", err)
	}

	// Verify the update
	updatedEmail, err := GetSubmissionEmailByID(db, createdEmail.ID)
	if err != nil {
		t.Fatalf("Failed to get updated email: %v", err)
	}

	if updatedEmail.Status != "failed" {
		t.Errorf("Expected status 'failed', got '%s'", updatedEmail.Status)
	}

	if updatedEmail.ErrorMessage != "SMTP connection failed" {
		t.Errorf("Expected error message 'SMTP connection failed', got '%s'", updatedEmail.ErrorMessage)
	}
}