package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"staticsend/pkg/email"
	"staticsend/pkg/models"
	"staticsend/pkg/turnstile"
)

// SubmissionHandler handles form submission requests
type SubmissionHandler struct {
	DB          *sql.DB
	EmailService *email.EmailService
}

// NewSubmissionHandler creates a new submission handler
func NewSubmissionHandler(db *sql.DB, emailService *email.EmailService) *SubmissionHandler {
	return &SubmissionHandler{
		DB:          db,
		EmailService: emailService,
	}
}

// SubmitForm handles form submissions
func (h *SubmissionHandler) SubmitForm(w http.ResponseWriter, r *http.Request) {
	// Get form key from URL path
	formKey := strings.TrimPrefix(r.URL.Path, "/api/v1/submit/")
	if formKey == "" {
		http.Error(w, "Form key is required", http.StatusBadRequest)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Get Turnstile token
	turnstileToken := r.FormValue("cf-turnstile-response")
	if turnstileToken == "" {
		http.Error(w, "Turnstile verification required", http.StatusBadRequest)
		return
	}

	// Get form from database
	form, err := models.GetFormByKey(h.DB, formKey)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if form == nil {
		http.Error(w, "Form not found", http.StatusNotFound)
		return
	}

	// Validate Turnstile token
	validator := turnstile.NewValidator(form.TurnstileSecret)
	remoteIP := getClientIP(r)
	
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	
	verification, err := validator.Verify(ctx, turnstileToken, remoteIP)
	if err != nil {
		http.Error(w, "Turnstile verification failed", http.StatusInternalServerError)
		return
	}
	
	if !verification.IsValid() {
		http.Error(w, "Invalid Turnstile token", http.StatusBadRequest)
		return
	}

	// Extract form data (excluding Turnstile token)
	formData := make(map[string]string)
	for key, values := range r.Form {
		if key != "cf-turnstile-response" && len(values) > 0 {
			formData[key] = values[0]
		}
	}

	// Convert form data to JSON for storage
	formDataJSON, err := json.Marshal(formData)
	if err != nil {
		http.Error(w, "Failed to process form data", http.StatusInternalServerError)
		return
	}

	// Create submission record
	userAgent := r.UserAgent()
	submission, err := models.CreateSubmission(h.DB, form.ID, remoteIP, userAgent, formDataJSON)
	if err != nil {
		http.Error(w, "Failed to save submission", http.StatusInternalServerError)
		return
	}

	// Send email notification asynchronously
	go func() {
		if err := h.EmailService.SendFormSubmissionAsync([]string{form.ForwardEmail}, formData); err != nil {
			// Log error but don't fail the request
			fmt.Printf("Failed to queue email: %v\n", err)
			// Update submission status to failed
			models.UpdateSubmissionStatus(h.DB, submission.ID, "failed")
		} else {
			// Update submission status to processed
			models.UpdateSubmissionStatus(h.DB, submission.ID, "processed")
		}
	}()

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Form submitted successfully",
		"submission_id": submission.ID,
	})
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (for proxies)
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		if ips := strings.Split(forwarded, ","); len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Fall back to remote address
	return strings.Split(r.RemoteAddr, ":")[0]
}