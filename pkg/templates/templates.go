package templates

import (
	"encoding/json"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"staticsend/pkg/models"
)

// TemplateData holds data for template rendering
type TemplateData struct {
	Title                  string
	User                   *models.User
	Error                  string
	Flash                  string
	ShowHeader             bool
	Forms                  []*models.Form
	Stats                  *DashboardStats
	Data                   interface{} // Generic data field for additional data
	AuthTurnstilePublicKey string      // Turnstile public key for auth pages
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
	baseURL   string
}

// NewTemplateManager creates a new template manager
func NewTemplateManager() *TemplateManager {
	tm := &TemplateManager{
		templates: make(map[string]*template.Template),
		baseURL:   getBaseURL(),
	}
	tm.loadTemplates()
	return tm
}

// templateFuncMap returns the template function map
func (tm *TemplateManager) templateFuncMap() template.FuncMap {
	return template.FuncMap{
		"unmarshalJSON": func(s string) (map[string]interface{}, error) {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(s), &data); err != nil {
				return nil, err
			}
			return data, nil
		},
		"baseURL": func() string {
			return tm.baseURL
		},
	}
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

	// Parse base template first with functions
	basePath := filepath.Join(cwd, "templates", "base.html")
	baseTmpl := template.Must(template.New("base.html").Funcs(tm.templateFuncMap()).ParseFiles(basePath))

	// Walk through all template files
	templatesDir := filepath.Join(cwd, "templates")
	
	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".html" && path != basePath {
			// Use relative path from templates directory as key
			relPath, _ := filepath.Rel(templatesDir, path)
			
			// Check if this is a partial (in partials directory)
			if filepath.Dir(relPath) == "partials" {
				// For partials, parse without base template but with functions
				tmpl := template.Must(template.New(filepath.Base(path)).Funcs(tm.templateFuncMap()).ParseFiles(path))
				tm.templates[relPath] = tmpl
			} else {
				// For full pages, use base template wrapper with functions
				tmpl := template.Must(baseTmpl.Clone())
				tmpl = template.Must(tmpl.Funcs(tm.templateFuncMap()).ParseFiles(path))
				tm.templates[relPath] = tmpl
			}
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

// getBaseURL determines the base URL for the application
func getBaseURL() string {
	// Try to get from environment variable
	if envURL := os.Getenv("STATICSEND_BASE_URL"); envURL != "" {
		return strings.TrimSuffix(envURL, "/")
	}
	
	// For development, use localhost with default port
	return "http://localhost:8080"
}