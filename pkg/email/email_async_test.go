package email

import (
	"strings"
	"testing"
	"time"
)

func TestSendAsync(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	service := NewEmailService(config, 10, 1, 2)
	defer service.Shutdown()

	// Test async send - should return immediately
	err := service.SendAsync([]string{"test@example.com"}, "Test Subject", "Test body")
	if err != nil {
		t.Errorf("SendAsync should not return error immediately: %v", err)
	}

	// Queue should have one item
	if service.QueueSize() != 1 {
		t.Errorf("Expected queue size 1, got %d", service.QueueSize())
	}

	// Give the worker a moment to process
	time.Sleep(100 * time.Millisecond)

	// Queue should be empty now (even though the send will fail)
	if service.QueueSize() != 0 {
		t.Errorf("Expected queue size 0 after processing, got %d", service.QueueSize())
	}
}

func TestSendAsync_QueueFull(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	// Create service with very small queue
	service := NewEmailService(config, 1, 1, 1)
	defer service.Shutdown()

	// Fill the queue
	err := service.SendAsync([]string{"test1@example.com"}, "Test 1", "Body 1")
	if err != nil {
		t.Errorf("First SendAsync should succeed: %v", err)
	}

	// Try to add another - should fail with queue full
	err = service.SendAsync([]string{"test2@example.com"}, "Test 2", "Body 2")
	if err == nil {
		t.Error("Second SendAsync should fail with queue full")
	}

	if !strings.Contains(err.Error(), "queue is full") {
		t.Errorf("Expected 'queue is full' error, got: %v", err)
	}
}

func TestSendFormSubmissionAsync(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	service := NewEmailService(config, 10, 1, 2)
	defer service.Shutdown()

	formData := map[string]string{
		"name":    "John Doe",
		"email":   "john@example.com",
		"message": "Test message",
	}

	// Test async form submission
	err := service.SendFormSubmissionAsync([]string{"admin@example.com"}, formData)
	if err != nil {
		t.Errorf("SendFormSubmissionAsync should not return error immediately: %v", err)
	}

	// Give the goroutine time to process the queue
	time.Sleep(10 * time.Millisecond)
	
	// Queue should have one item (or be processed already)
	queueSize := service.QueueSize()
	if queueSize != 1 && queueSize != 0 {
		t.Errorf("Expected queue size 1 or 0 (if processed), got %d", queueSize)
	}
}

func TestShutdown(t *testing.T) {
	config := EmailConfig{
		Host:     "smtp.example.com",
		Port:     587,
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
		UseTLS:   true,
	}

	service := NewEmailService(config, 10, 2, 2)

	// Add some jobs to the queue
	for i := 0; i < 3; i++ {
		err := service.SendAsync([]string{"test@example.com"}, "Test", "Body")
		if err != nil {
			t.Errorf("SendAsync should succeed: %v", err)
		}
	}

	// Shutdown should complete without hanging
	service.Shutdown()

	// After shutdown, SendAsync should fail
	err := service.SendAsync([]string{"test@example.com"}, "Test", "Body")
	if err == nil {
		t.Error("SendAsync should fail after shutdown")
	}
}