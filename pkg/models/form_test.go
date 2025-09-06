package models

import (
	"testing"
)

func TestCreateForm(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user first
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test creating a new form
	form, err := CreateForm(db, user.ID, "contact", "Contact Form", "Get in touch with us", "https://example.com/thank-you")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	if form.Name != "contact" {
		t.Errorf("Expected name 'contact', got '%s'", form.Name)
	}

	if form.Title != "Contact Form" {
		t.Errorf("Expected title 'Contact Form', got '%s'", form.Title)
	}

	if form.UserID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, form.UserID)
	}

	// Test duplicate form name for same user
	_, err = CreateForm(db, user.ID, "contact", "Another Form", "Different form", "")
	if err == nil {
		t.Error("Expected error when creating form with duplicate name for same user")
	}

	// Test same form name for different user (should work)
	user2, err := CreateUser(db, "user2@example.com", "hashed_password2")
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	form2, err := CreateForm(db, user2.ID, "contact", "Contact Form", "Second user's form", "")
	if err != nil {
		t.Fatalf("Failed to create form for second user: %v", err)
	}

	if form2.Name != "contact" {
		t.Errorf("Expected name 'contact' for second user, got '%s'", form2.Name)
	}
}

func TestGetFormByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test data
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	createdForm, err := CreateForm(db, user.ID, "contact", "Contact Form", "Test form", "")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Test retrieving the form
	retrievedForm, err := GetFormByID(db, createdForm.ID)
	if err != nil {
		t.Fatalf("Failed to get form by ID: %v", err)
	}

	if retrievedForm == nil {
		t.Fatal("Expected to retrieve form, got nil")
	}

	if retrievedForm.Name != "contact" {
		t.Errorf("Expected name 'contact', got '%s'", retrievedForm.Name)
	}

	// Test non-existent form
	nonExistentForm, err := GetFormByID(db, 999)
	if err != nil {
		t.Fatalf("Unexpected error getting non-existent form: %v", err)
	}

	if nonExistentForm != nil {
		t.Error("Expected nil for non-existent form")
	}
}

func TestGetFormsByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test users
	user1, err := CreateUser(db, "user1@example.com", "hashed_password1")
	if err != nil {
		t.Fatalf("Failed to create user1: %v", err)
	}

	user2, err := CreateUser(db, "user2@example.com", "hashed_password2")
	if err != nil {
		t.Fatalf("Failed to create user2: %v", err)
	}

	// Create forms for user1
	_, err = CreateForm(db, user1.ID, "contact", "Contact Form", "User1 form 1", "")
	if err != nil {
		t.Fatalf("Failed to create form 1 for user1: %v", err)
	}

	_, err = CreateForm(db, user1.ID, "feedback", "Feedback Form", "User1 form 2", "")
	if err != nil {
		t.Fatalf("Failed to create form 2 for user1: %v", err)
	}

	// Create form for user2
	_, err = CreateForm(db, user2.ID, "contact", "Contact Form", "User2 form", "")
	if err != nil {
		t.Fatalf("Failed to create form for user2: %v", err)
	}

	// Test getting forms for user1
	user1Forms, err := GetFormsByUserID(db, user1.ID)
	if err != nil {
		t.Fatalf("Failed to get forms for user1: %v", err)
	}

	if len(user1Forms) != 2 {
		t.Errorf("Expected 2 forms for user1, got %d", len(user1Forms))
	}

	// Test getting forms for user2
	user2Forms, err := GetFormsByUserID(db, user2.ID)
	if err != nil {
		t.Fatalf("Failed to get forms for user2: %v", err)
	}

	if len(user2Forms) != 1 {
		t.Errorf("Expected 1 form for user2, got %d", len(user2Forms))
	}

	// Test getting forms for non-existent user
	nonExistentForms, err := GetFormsByUserID(db, 999)
	if err != nil {
		t.Fatalf("Unexpected error getting forms for non-existent user: %v", err)
	}

	if len(nonExistentForms) != 0 {
		t.Errorf("Expected 0 forms for non-existent user, got %d", len(nonExistentForms))
	}
}

func TestFormExists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test user and form
	user, err := CreateUser(db, "user@example.com", "hashed_password")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	_, err = CreateForm(db, user.ID, "contact", "Contact Form", "Test form", "")
	if err != nil {
		t.Fatalf("Failed to create form: %v", err)
	}

	// Test existing form
	exists, err := FormExists(db, user.ID, "contact")
	if err != nil {
		t.Fatalf("Failed to check form existence: %v", err)
	}

	if !exists {
		t.Error("Expected form to exist")
	}

	// Test non-existent form name
	exists, err = FormExists(db, user.ID, "nonexistent")
	if err != nil {
		t.Fatalf("Failed to check non-existent form: %v", err)
	}

	if exists {
		t.Error("Expected form to not exist")
	}

	// Test form exists for different user (should return false)
	user2, err := CreateUser(db, "user2@example.com", "hashed_password2")
	if err != nil {
		t.Fatalf("Failed to create second user: %v", err)
	}

	exists, err = FormExists(db, user2.ID, "contact")
	if err != nil {
		t.Fatalf("Failed to check form existence for different user: %v", err)
	}

	if exists {
		t.Error("Expected form to not exist for different user")
	}
}