package web

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"staticsend/pkg/middleware"
	"staticsend/pkg/models"
	"staticsend/pkg/templates"
)

// WebHandler handles web page requests
type WebHandler struct {
	DB                     *sql.DB
	TemplateManager        *templates.TemplateManager
	AuthTurnstilePublicKey string
}

// NewWebHandler creates a new web handler
func NewWebHandler(db *sql.DB, tm *templates.TemplateManager, authTurnstilePublicKey string) *WebHandler {
	return &WebHandler{
		DB:                     db,
		TemplateManager:        tm,
		AuthTurnstilePublicKey: authTurnstilePublicKey,
	}
}



// LoginPage renders the login page
func (h *WebHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := templates.TemplateData{
		Title:                  "Login - staticSend",
		ShowHeader:             false,
		AuthTurnstilePublicKey: h.AuthTurnstilePublicKey,
	}
	
	if err := h.TemplateManager.Render(w, "auth/login.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// RegisterPage renders the registration page
func (h *WebHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := templates.TemplateData{
		Title:                  "Register - staticSend",
		ShowHeader:             false,
		AuthTurnstilePublicKey: h.AuthTurnstilePublicKey,
	}
	
	if err := h.TemplateManager.Render(w, "auth/register.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// Dashboard renders the main dashboard
func (h *WebHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// Fetch user's forms from database
	forms, err := models.GetFormsByUserID(h.DB, user.ID)
	if err != nil {
		http.Error(w, "Failed to fetch forms", http.StatusInternalServerError)
		return
	}

	// Convert to pointer slice for template
	formPtrs := make([]*models.Form, len(forms))
	for i := range forms {
		formPtrs[i] = &forms[i]
	}

	// Get submission count for each form
	for _, form := range formPtrs {
		count, err := models.GetSubmissionCountByFormID(h.DB, form.ID)
		if err == nil {
			form.SubmissionCount = count
		}
	}

	// Get total submission count
	totalSubmissions := 0
	for _, form := range formPtrs {
		totalSubmissions += form.SubmissionCount
	}

	data := templates.DefaultTemplateData()
	data.Title = "Dashboard - staticSend"
	data.User = user
	data.Forms = formPtrs
	data.Stats.FormCount = len(formPtrs)
	data.Stats.SubmissionCount = totalSubmissions

	if err := h.TemplateManager.Render(w, "dashboard/index.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// CreateFormModal renders the create form modal
func (h *WebHandler) CreateFormModal(w http.ResponseWriter, r *http.Request) {
	data := templates.TemplateData{
		Title: "Create New Form",
	}
	
	// Render the partial for the modal content
	// HTMX will handle replacing the content in #modal-content
	// The button click already adds .overflow-hidden to body and shows the modal
	if err := h.TemplateManager.Render(w, "partials/form_modal.html", data); err != nil {
		// Log the specific error for debugging
		log.Printf("Failed to render form modal template: %v", err)
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// ViewFormModal renders the view form modal
func (h *WebHandler) ViewFormModal(w http.ResponseWriter, r *http.Request) {
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

	// Fetch form from database
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

	data := templates.TemplateData{
		Title: "View Form - " + form.Name,
		Data:  form,
	}
	
	if err := h.TemplateManager.Render(w, "partials/view_form_modal.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// EditFormModal renders the edit form modal
func (h *WebHandler) EditFormModal(w http.ResponseWriter, r *http.Request) {
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

	// Fetch form from database
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

	data := templates.TemplateData{
		Title: "Edit Form - " + form.Name,
		Data:  form,
	}
	
	if err := h.TemplateManager.Render(w, "partials/edit_form_modal.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// FormSubmissions renders the form submissions page
func (h *WebHandler) FormSubmissions(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	formIDStr := chi.URLParam(r, "id")
	formID, err := strconv.ParseInt(formIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid form ID", http.StatusBadRequest)
		return
	}

	// Fetch form from database
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

	// Get submissions for this form
	submissions, err := models.GetSubmissionsByFormID(h.DB, form.ID)
	if err != nil {
		http.Error(w, "Failed to fetch submissions", http.StatusInternalServerError)
		return
	}

	// Get submission count
	count, err := models.GetSubmissionCountByFormID(h.DB, form.ID)
	if err == nil {
		form.SubmissionCount = count
	}

	data := templates.DefaultTemplateData()
	data.Title = "Submissions - " + form.Name + " - staticSend"
	data.User = user
	data.Data = map[string]interface{}{
		"Form":        form,
		"Submissions": submissions,
	}

	if err := h.TemplateManager.Render(w, "submissions/index.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}