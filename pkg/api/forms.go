package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"staticsend/pkg/middleware"
	"staticsend/pkg/models"
)

// FormHandler handles form-related API requests
type FormHandler struct {
	DB *sql.DB
}

// NewFormHandler creates a new form handler
func NewFormHandler(db *sql.DB) *FormHandler {
	return &FormHandler{
		DB: db,
	}
}

// CreateForm handles form creation
func (h *FormHandler) CreateForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	title := r.FormValue("title")
	description := r.FormValue("description")
	redirectURL := r.FormValue("redirect_url")

	if name == "" || title == "" {
		http.Error(w, "Name and title are required", http.StatusBadRequest)
		return
	}

	// Check if form name already exists for this user
	exists, err := models.FormExists(h.DB, user.ID, name)
	if err != nil {
		http.Error(w, "Failed to check form existence", http.StatusInternalServerError)
		return
	}
	if exists {
		http.Error(w, "Form with this name already exists", http.StatusConflict)
		return
	}

	form, err := models.CreateForm(h.DB, user.ID, name, title, description, redirectURL)
	if err != nil {
		http.Error(w, "Failed to create form", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(form)
}

// GetForm handles retrieving a single form
func (h *FormHandler) GetForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	formIDStr := chi.URLParam(r, "id")
	formID, err := strconv.ParseInt(formIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid form ID", http.StatusBadRequest)
		return
	}

	form, err := models.GetFormByID(h.DB, formID)
	if err != nil {
		http.Error(w, "Failed to fetch form", http.StatusInternalServerError)
		return
	}
	if form == nil {
		http.Error(w, "Form not found", http.StatusNotFound)
		return
	}

	// Verify user owns this form
	if form.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get submission count
	count, err := models.GetSubmissionCountByFormID(h.DB, form.ID)
	if err == nil {
		form.SubmissionCount = count
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(form)
}

// DeleteForm handles form deletion
func (h *FormHandler) DeleteForm(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	formIDStr := chi.URLParam(r, "id")
	formID, err := strconv.ParseInt(formIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid form ID", http.StatusBadRequest)
		return
	}

	form, err := models.GetFormByID(h.DB, formID)
	if err != nil {
		http.Error(w, "Failed to fetch form", http.StatusInternalServerError)
		return
	}
	if form == nil {
		http.Error(w, "Form not found", http.StatusNotFound)
		return
	}

	// Verify user owns this form
	if form.UserID != user.ID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete form from database
	_, err = h.DB.Exec("DELETE FROM forms WHERE id = ?", formID)
	if err != nil {
		http.Error(w, "Failed to delete form", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetUserForms handles retrieving all forms for a user
func (h *FormHandler) GetUserForms(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	forms, err := models.GetFormsByUserID(h.DB, user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch forms", http.StatusInternalServerError)
		return
	}

	// Get submission counts for each form
	formPtrs := make([]*models.Form, len(forms))
	for i := range forms {
		formPtrs[i] = &forms[i]
		count, err := models.GetSubmissionCountByFormID(h.DB, formPtrs[i].ID)
		if err == nil {
			formPtrs[i].SubmissionCount = count
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formPtrs)
}