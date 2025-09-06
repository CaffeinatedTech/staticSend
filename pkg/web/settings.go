package web

import (
	"encoding/json"
	"net/http"

	"staticsend/pkg/database"
	"staticsend/pkg/models"
	"staticsend/pkg/templates"
)

// SettingsHandler handles application settings
type SettingsHandler struct {
	DB        *database.Database
	Templates *templates.TemplateManager
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(db *database.Database, tm *templates.TemplateManager) *SettingsHandler {
	return &SettingsHandler{
		DB:        db,
		Templates: tm,
	}
}

// SettingsPage renders the settings page
func (h *SettingsHandler) SettingsPage(w http.ResponseWriter, r *http.Request) {
	settings, err := models.GetAllAppSettings(h.DB.Connection)
	if err != nil {
		h.renderSettingsPage(w, "Failed to load settings", nil)
		return
	}

	h.renderSettingsPage(w, "", settings)
}

// UpdateSettings handles updating application settings
func (h *SettingsHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderSettingsPage(w, "Invalid form data", nil)
		return
	}

	// Handle checkbox settings specifically - registration_enabled
	// The hidden field ensures we always get a value ("false" when unchecked, "true" when checked)
	if registrationEnabled := r.FormValue("registration_enabled"); registrationEnabled != "" {
		if err := models.UpdateAppSetting(h.DB.Connection, "registration_enabled", registrationEnabled); err != nil {
			h.renderSettingsPage(w, "Failed to update registration setting", nil)
			return
		}
	}

	// Handle text settings - only update if provided
	if siteTitle := r.FormValue("site_title"); siteTitle != "" {
		if err := models.UpdateAppSetting(h.DB.Connection, "site_title", siteTitle); err != nil {
			h.renderSettingsPage(w, "Failed to update site title", nil)
			return
		}
	}

	if siteDescription := r.FormValue("site_description"); siteDescription != "" {
		if err := models.UpdateAppSetting(h.DB.Connection, "site_description", siteDescription); err != nil {
			h.renderSettingsPage(w, "Failed to update site description", nil)
			return
		}
	}

	// Redirect back to dashboard after saving
	w.Header().Set("HX-Redirect", "/dashboard")
}

// GetRegistrationStatus returns the current registration status as JSON
func (h *SettingsHandler) GetRegistrationStatus(w http.ResponseWriter, r *http.Request) {
	enabled, err := models.IsRegistrationEnabled(h.DB.Connection)
	if err != nil {
		http.Error(w, "Failed to get registration status", http.StatusInternalServerError)
		return
	}

	response := map[string]bool{"enabled": enabled}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// renderSettingsPage renders the settings page with an optional error
func (h *SettingsHandler) renderSettingsPage(w http.ResponseWriter, errorMsg string, settings []models.AppSetting) {
	data := templates.TemplateData{
		Title:      "Settings - staticSend",
		Error:      errorMsg,
		ShowHeader: true,
		Data:       settings,
	}

	h.Templates.Render(w, "settings/index.html", data)
}