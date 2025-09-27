package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"staticsend/pkg/middleware"
	"staticsend/pkg/models"
	"staticsend/pkg/templates"
)

// Test that the submissions page renders without error and properly
// uses the unmarshalJSON helper to display fields from SubmittedData.
func TestFormSubmissions_RendersSubmittedData(t *testing.T) {
	// 1) Setup DB with schema and seed minimal data
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	user, err := models.CreateUser(db, "test@example.com", "hashedpass")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	form, err := models.CreateForm(db, user.ID, "Contact", "example.com", "secret", "dest@example.com", "", "formkey-123")
	if err != nil {
		t.Fatalf("CreateForm: %v", err)
	}

	// Insert a couple of submissions with JSON payloads
	payload1, _ := json.Marshal(map[string]string{"name": "Alice", "email": "alice@example.com"})
	if _, err := models.CreateSubmission(db, form.ID, "203.0.113.10", "UA1", payload1); err != nil {
		t.Fatalf("CreateSubmission #1: %v", err)
	}
	payload2, _ := json.Marshal(map[string]string{"name": "Bob", "email": "bob@example.com"})
	if _, err := models.CreateSubmission(db, form.ID, "203.0.113.20", "UA2", payload2); err != nil {
		t.Fatalf("CreateSubmission #2: %v", err)
	}

	// 2) Create a temporary templates dir matching the loader expectation (cwd/templates/...)
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })

	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("Chdir: %v", err)
	}

	// Minimal base.html that includes the content block
	base := `<!DOCTYPE html><html><body>{{template "content" .}}</body></html>`
	if err := os.MkdirAll(filepath.Join("templates", "submissions"), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join("templates", "base.html"), []byte(base), 0o644); err != nil {
		t.Fatalf("write base.html: %v", err)
	}

	// Submissions page that uses the helper pipeline similar to real template
	subsTpl := `{{define "content"}}{{range .Data.Submissions}}{{$m := .SubmittedData | unmarshalJSON}}{{index $m "name"}} {{end}}{{end}}`
	if err := os.WriteFile(filepath.Join("templates", "submissions", "index.html"), []byte(subsTpl), 0o644); err != nil {
		t.Fatalf("write submissions/index.html: %v", err)
	}

	// 3) Build TemplateManager and WebHandler
	tm := templates.NewTemplateManager()
	h := NewWebHandler(db, tm, "")

	// 4) Build a request with chi route params and user in context
	req := httptest.NewRequest(http.MethodGet, "/forms/"+strconv.FormatInt(form.ID, 10)+"/submissions", nil)
	// chi URLParam is used inside handler; add route context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", strconv.FormatInt(form.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	// add user to context (bypass auth middleware)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserKey, user))

	rec := httptest.NewRecorder()
	h.FormSubmissions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got=%d body=%s", rec.Code, rec.Body.String())
	}

	out := rec.Body.String()
	if !containsAll(out, []string{"Alice", "Bob"}) {
		t.Fatalf("expected rendered output to contain submission names; got: %q", out)
	}
}

// containsAll returns true if s contains all substrings in subs
func containsAll(s string, subs []string) bool {
	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			return false
		}
	}
	return true
}
