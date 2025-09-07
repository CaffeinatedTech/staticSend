package templates

import (
	"testing"
)

func TestTemplateData_Fields(t *testing.T) {
	data := TemplateData{
		Title:                  "Test Title",
		Error:                  "Test Error",
		Flash:                  "Test Flash",
		ShowHeader:             true,
		AuthTurnstilePublicKey: "test-key",
	}

	if data.Title != "Test Title" {
		t.Errorf("Expected Title 'Test Title', got '%s'", data.Title)
	}

	if data.Error != "Test Error" {
		t.Errorf("Expected Error 'Test Error', got '%s'", data.Error)
	}

	if data.Flash != "Test Flash" {
		t.Errorf("Expected Flash 'Test Flash', got '%s'", data.Flash)
	}

	if !data.ShowHeader {
		t.Error("Expected ShowHeader to be true")
	}

	if data.AuthTurnstilePublicKey != "test-key" {
		t.Errorf("Expected AuthTurnstilePublicKey 'test-key', got '%s'", data.AuthTurnstilePublicKey)
	}
}

func TestDefaultTemplateData(t *testing.T) {
	data := DefaultTemplateData()

	if data.Title != "staticSend" {
		t.Errorf("Expected Title 'staticSend', got '%s'", data.Title)
	}

	if !data.ShowHeader {
		t.Error("Expected ShowHeader to be true")
	}

	if data.Stats == nil {
		t.Error("Expected Stats to be initialized")
	}

	if data.Stats.FormCount != 0 {
		t.Errorf("Expected FormCount 0, got %d", data.Stats.FormCount)
	}

	if data.Stats.SubmissionCount != 0 {
		t.Errorf("Expected SubmissionCount 0, got %d", data.Stats.SubmissionCount)
	}
}

func TestDashboardStats(t *testing.T) {
	stats := &DashboardStats{
		FormCount:       5,
		SubmissionCount: 10,
	}

	if stats.FormCount != 5 {
		t.Errorf("Expected FormCount 5, got %d", stats.FormCount)
	}

	if stats.SubmissionCount != 10 {
		t.Errorf("Expected SubmissionCount 10, got %d", stats.SubmissionCount)
	}
}
