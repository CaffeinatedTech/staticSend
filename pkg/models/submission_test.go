package models

import (
	"encoding/json"
	"testing"
)

func TestCreateSubmission(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test user and form
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form := CreateTestForm(t, db, user.ID, "contact", "example.com", "turnstile_secret_456", "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Test creating a new submission
	submissionData := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"message": "Hello, this is a test message",
	}
	
	dataBytes, err := json.Marshal(submissionData)
	if err != nil {
		t.Fatalf("Failed to marshal submission data: %v", err)
	}

	submission, err := CreateSubmission(db, form.ID, "192.168.1.1", "Test Browser", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	if submission.FormID != form.ID {
		t.Errorf("Expected form ID %d, got %d", form.ID, submission.FormID)
	}

	if submission.IPAddress != "192.168.1.1" {
		t.Errorf("Expected IP address '192.168.1.1', got '%s'", submission.IPAddress)
	}

	if submission.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", submission.Status)
	}

	if submission.ProcessedAt != nil {
		t.Error("Expected processed_at to be nil for new submission")
	}

	// Verify the submitted data
	var retrievedData map[string]interface{}
	if err := json.Unmarshal(submission.SubmittedData, &retrievedData); err != nil {
		t.Fatalf("Failed to unmarshal submitted data: %v", err)
	}

	if retrievedData["name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%v'", retrievedData["name"])
	}
}

func TestGetSubmissionByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form := CreateTestForm(t, db, user.ID, "contact", "example.com", "turnstile_secret_456", "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	submissionData := map[string]interface{}{"test": "data"}
	dataBytes, _ := json.Marshal(submissionData)
	createdSubmission, err := CreateSubmission(db, form.ID, "192.168.1.1", "Test Browser", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	// Test retrieving the submission
	retrievedSubmission, err := GetSubmissionByID(db, createdSubmission.ID)
	if err != nil {
		t.Fatalf("Failed to get submission by ID: %v", err)
	}

	if retrievedSubmission == nil {
		t.Fatal("Expected to retrieve submission, got nil")
	}

	if retrievedSubmission.ID != createdSubmission.ID {
		t.Errorf("Expected ID %d, got %d", createdSubmission.ID, retrievedSubmission.ID)
	}

	// Test non-existent submission
	nonExistentSubmission, err := GetSubmissionByID(db, 999)
	if err != nil {
		t.Fatalf("Unexpected error getting non-existent submission: %v", err)
	}

	if nonExistentSubmission != nil {
		t.Error("Expected nil for non-existent submission")
	}
}

func TestGetSubmissionsByFormID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form1, err := CreateForm(db, user.ID, "contact", "example1.com", "turnstile_secret_456", "admin@example.com", "form_key_001")
	if err != nil {
		t.Fatalf("Failed to create form1: %v", err)
	}

	form2, err := CreateForm(db, user.ID, "feedback", "example1.com", "turnstile_secret_abc", "admin@example.com", "form_key_002")
	if err != nil {
		t.Fatalf("Failed to create form2: %v", err)
	}

	// Create submissions for form1
	submissionData := map[string]interface{}{"test": "data"}
	dataBytes, _ := json.Marshal(submissionData)
	
	_, err = CreateSubmission(db, form1.ID, "192.168.1.1", "Browser 1", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission 1 for form1: %v", err)
	}

	_, err = CreateSubmission(db, form1.ID, "192.168.1.2", "Browser 2", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission 2 for form1: %v", err)
	}

	// Create submission for form2
	_, err = CreateSubmission(db, form2.ID, "192.168.1.3", "Browser 3", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission for form2: %v", err)
	}

	// Test getting submissions for form1
	form1Submissions, err := GetSubmissionsByFormID(db, form1.ID)
	if err != nil {
		t.Fatalf("Failed to get submissions for form1: %v", err)
	}

	if len(form1Submissions) != 2 {
		t.Errorf("Expected 2 submissions for form1, got %d", len(form1Submissions))
	}

	// Test getting submissions for form2
	form2Submissions, err := GetSubmissionsByFormID(db, form2.ID)
	if err != nil {
		t.Fatalf("Failed to get submissions for form2: %v", err)
	}

	if len(form2Submissions) != 1 {
		t.Errorf("Expected 1 submission for form2, got %d", len(form2Submissions))
	}

	// Test getting submissions for non-existent form
	nonExistentSubmissions, err := GetSubmissionsByFormID(db, 999)
	if err != nil {
		t.Fatalf("Unexpected error getting submissions for non-existent form: %v", err)
	}

	if len(nonExistentSubmissions) != 0 {
		t.Errorf("Expected 0 submissions for non-existent form, got %d", len(nonExistentSubmissions))
	}
}

func TestUpdateSubmissionStatus(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form := CreateTestForm(t, db, user.ID, "contact", "example.com", "turnstile_secret_456", "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	submissionData := map[string]interface{}{"test": "data"}
	dataBytes, _ := json.Marshal(submissionData)
	submission, err := CreateSubmission(db, form.ID, "192.168.1.1", "Test Browser", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission: %v", err)
	}

	// Test updating status to processed
	err = UpdateSubmissionStatus(db, submission.ID, "processed")
	if err != nil {
		t.Fatalf("Failed to update submission status: %v", err)
	}

	// Verify the update
	updatedSubmission, err := GetSubmissionByID(db, submission.ID)
	if err != nil {
		t.Fatalf("Failed to get updated submission: %v", err)
	}

	if updatedSubmission.Status != "processed" {
		t.Errorf("Expected status 'processed', got '%s'", updatedSubmission.Status)
	}

	if updatedSubmission.ProcessedAt == nil {
		t.Error("Expected processed_at to be set for processed submission")
	}

	// Test updating status back to pending
	err = UpdateSubmissionStatus(db, submission.ID, "pending")
	if err != nil {
		t.Fatalf("Failed to update submission status to pending: %v", err)
	}

	// Verify the update
	updatedSubmission, err = GetSubmissionByID(db, submission.ID)
	if err != nil {
		t.Fatalf("Failed to get updated submission: %v", err)
	}

	if updatedSubmission.Status != "pending" {
		t.Errorf("Expected status 'pending', got '%s'", updatedSubmission.Status)
	}

	if updatedSubmission.ProcessedAt != nil {
		t.Error("Expected processed_at to be nil for pending submission")
	}
}

func TestGetSubmissionCountByFormID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	form := CreateTestForm(t, db, user.ID, "contact", "example.com", "turnstile_secret_456", "admin@example.com")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Test count with no submissions
	count, err := GetSubmissionCountByFormID(db, form.ID)
	if err != nil {
		t.Fatalf("Failed to get submission count: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 submissions, got %d", count)
	}

	// Create submissions
	submissionData := map[string]interface{}{"test": "data"}
	dataBytes, _ := json.Marshal(submissionData)
	
	_, err = CreateSubmission(db, form.ID, "192.168.1.1", "Browser 1", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission 1: %v", err)
	}

	_, err = CreateSubmission(db, form.ID, "192.168.1.2", "Browser 2", dataBytes)
	if err != nil {
		t.Fatalf("Failed to create submission 2: %v", err)
	}

	// Test count with submissions
	count, err = GetSubmissionCountByFormID(db, form.ID)
	if err != nil {
		t.Fatalf("Failed to get submission count: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 submissions, got %d", count)
	}

	// Test count for non-existent form
	count, err = GetSubmissionCountByFormID(db, 999)
	if err != nil {
		t.Fatalf("Failed to get submission count for non-existent form: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 submissions for non-existent form, got %d", count)
	}
}