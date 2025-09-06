package templates

import (
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"

	"staticsend/pkg/models"
)

// TemplateData holds data for template rendering
type TemplateData struct {
	Title      string
	User       *models.User
	Error      string
	Flash      string
	ShowHeader bool
	Forms      []*models.Form
	Stats      *DashboardStats
	Data       interface{} // Generic data field for additional data
}

// DashboardStats holds statistics for the dashboard
type DashboardStats struct {
	FormCount       int
	SubmissionCount int
}

// TemplateManager handles template parsing and rendering
type TemplateManager struct {
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
	}
	tm.loadTemplates()
	return tm
}

// loadTemplates loads all templates from the templates directory
func (tm *TemplateManager) loadTemplates() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting working directory: %v", err)
		return
	}

	// Parse base template first
	basePath := filepath.Join(cwd, "templates", "base.html")
	baseTmpl := template.Must(template.ParseFiles(basePath))

	// Walk through all template files
	templatesDir := filepath.Join(cwd, "templates")
	
	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".html" && path != basePath {
			// Create a new template by cloning the base and adding the specific template
			tmpl := template.Must(baseTmpl.Clone())
			tmpl = template.Must(tmpl.ParseFiles(path))
			
			// Use relative path from templates directory as key
			relPath, _ := filepath.Rel(templatesDir, path)
			tm.templates[relPath] = tmpl
		}
		return nil
	})

	if err != nil {
		log.Printf("Error loading templates: %v", err)
	}
}

// Render renders a template with the given data
func (tm *TemplateManager) Render(w io.Writer, name string, data TemplateData) error {
	tm.mu.RLock()
	tmpl, exists := tm.templates[name]
	tm.mu.RUnlock()

	if !exists {
		// Try to reload templates if not found
		tm.loadTemplates()
		tm.mu.RLock()
		tmpl, exists = tm.templates[name]
		tm.mu.RUnlock()

		if !exists {
			return os.ErrNotExist
		}
	}

	return tmpl.Execute(w, data)
}

// DefaultTemplateData creates default template data with common values
func DefaultTemplateData() TemplateData {
	return TemplateData{
		Title:      "staticSend",
		ShowHeader: true,
		Stats: &DashboardStats{
			FormCount:       0,
			SubmissionCount: 0,
		},
	}
}