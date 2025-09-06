package web

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"staticsend/pkg/middleware"
	"staticsend/pkg/models"
	"staticsend/pkg/templates"
)

// WebHandler handles web page requests
type WebHandler struct {
	TemplateManager *templates.TemplateManager
}

// NewWebHandler creates a new web handler
func NewWebHandler(tm *templates.TemplateManager) *WebHandler {
	return &WebHandler{
		TemplateManager: tm,
	}
}



// LoginPage renders the login page
func (h *WebHandler) LoginPage(w http.ResponseWriter, r *http.Request) {
	data := templates.TemplateData{
		Title:      "Login - staticSend",
		ShowHeader: false,
	}
	
	if err := h.TemplateManager.Render(w, "auth/login.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// RegisterPage renders the registration page
func (h *WebHandler) RegisterPage(w http.ResponseWriter, r *http.Request) {
	data := templates.TemplateData{
		Title:      "Register - staticSend",
		ShowHeader: false,
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

	// TODO: Fetch user's forms and stats from database
	data := templates.DefaultTemplateData()
	data.Title = "Dashboard - staticSend"
	data.User = user
	data.Forms = []*models.Form{} // Empty for now
	data.Stats.FormCount = 0
	data.Stats.SubmissionCount = 0

	if err := h.TemplateManager.Render(w, "dashboard/index.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// CreateFormModal renders the create form modal
func (h *WebHandler) CreateFormModal(w http.ResponseWriter, r *http.Request) {
	data := templates.TemplateData{
		Title: "Create New Form",
	}
	
	// This would render a partial for the modal content
	w.Header().Set("HX-Trigger", "modalOpen")
	if err := h.TemplateManager.Render(w, "partials/form_modal.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}

// ViewFormModal renders the view form modal
func (h *WebHandler) ViewFormModal(w http.ResponseWriter, r *http.Request) {
	formIDStr := chi.URLParam(r, "id")
	_, err := strconv.ParseInt(formIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid form ID", http.StatusBadRequest)
		return
	}

	// TODO: Fetch form from database using formID
	data := templates.TemplateData{
		Title: "View Form",
	}
	
	w.Header().Set("HX-Trigger", "modalOpen")
	if err := h.TemplateManager.Render(w, "partials/view_form_modal.html", data); err != nil {
		http.Error(w, "Failed to render template", http.StatusInternalServerError)
	}
}