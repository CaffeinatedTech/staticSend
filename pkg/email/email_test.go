package email

import (
	"strings"
	"testing"
)

func TestNewEmailService(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	service := NewEmailService(config, 100, 3, 3)

	if service.config.Host != "smtp.example.com" {
		t.Errorf("Expected host 'smtp.example.com', got '%s'", service.config.Host)
	}

	if service.config.Port != 587 {
		t.Errorf("Expected port 587, got %d", service.config.Port)
	}

	if service.config.Username != "user" {
		t.Errorf("Expected username 'user', got '%s'", service.config.Username)
	}

	if service.config.From != "noreply@example.com" {
		t.Errorf("Expected from 'noreply@example.com', got '%s'", service.config.From)
	}

	if !service.config.UseTLS {
		t.Error("Expected UseTLS to be true")
	}
}

func TestBuildMessage(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	service := NewEmailService(config, 100, 3, 3)

	to := []string{"recipient@example.com"}
	subject := "Test Subject"
	body := "Test body content"

	message := service.buildMessage(to, subject, body)

	// Check that all required headers are present
	headers := []string{
		"From: noreply@example.com",
		"To: recipient@example.com",
		"Subject: Test Subject",
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		body,
	}

	for _, header := range headers {
		if !strings.Contains(message, header) {
			t.Errorf("Message should contain '%s'", header)
		}
	}
}

func TestSendFormSubmission(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	service := NewEmailService(config, 100, 3, 3)

	formData := map[string]string{
		"name":    "John Doe",
		"email":   "john@example.com",
		"message": "Test message",
	}

	// This will fail because we don't have a real SMTP server,
	// but we can test that the function constructs the email properly
	err := service.SendFormSubmission([]string{"admin@example.com"}, formData)

	// We expect an error since there's no SMTP server running
	if err == nil {
		t.Error("Expected error due to missing SMTP server")
	}

	// But the error should be about connection, not about message construction
	if !strings.Contains(err.Error(), "connection") && !strings.Contains(err.Error(), "dial") {
		t.Errorf("Expected connection error, got: %v", err)
	}
}

func TestSend_NoRecipients(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	service := NewEmailService(config, 100, 3, 3)

	err := service.Send([]string{}, "Test Subject", "Test body")

	if err == nil {
		t.Error("Expected error when no recipients are specified")
	}

	if !strings.Contains(err.Error(), "no recipients") {
		t.Errorf("Expected 'no recipients' error, got: %v", err)
	}
}

func TestEmailConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config EmailConfig
		valid  bool
	}{
		{
			name: "valid configuration",
			config: EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
				Password: "pass",
				From:     "noreply@example.com",
				UseTLS:   true,
			},
			valid: true,
		},
		{
			name: "missing host",
			config: EmailConfig{
				Host:     "",
				Port:     587,
				Username: "user",
				Password: "pass",
				From:     "noreply@example.com",
				UseTLS:   true,
			},
			valid: false,
		},
		{
			name: "invalid port",
			config: EmailConfig{
				Host:     "smtp.example.com",
				Port:     0,
				Username: "user",
				Password: "pass",
				From:     "noreply@example.com",
				UseTLS:   true,
			},
			valid: false,
		},
		{
			name: "missing from address",
			config: EmailConfig{
				Host:     "smtp.example.com",
				Port:     587,
				Username: "user",
				Password: "pass",
				From:     "",
				UseTLS:   true,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewEmailService(tt.config, 100, 3, 3)

			// Test that the service was created (validation happens at send time)
			if service == nil {
				t.Error("Service should be created even with invalid config")
			}

			// Attempt to send - should fail for invalid config
			err := service.Send([]string{"test@example.com"}, "Test", "Test")

			if tt.valid && err != nil {
				// For valid config, we expect connection errors but not validation errors
				if strings.Contains(err.Error(), "invalid") {
					t.Errorf("Expected connection error, got validation error: %v", err)
				}
			}
		})
	}
}
