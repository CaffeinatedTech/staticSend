package email

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"
)

// EmailConfig holds SMTP configuration
type EmailConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	UseTLS   bool
}

// EmailService handles email sending
type EmailService struct {
	config EmailConfig
}

// NewEmailService creates a new email service with the given configuration
func NewEmailService(config EmailConfig) *EmailService {
	return &EmailService{
		config: config,
	}
}

// Send sends an email with the given subject and body to the specified recipients
func (es *EmailService) Send(to []string, subject, body string) error {
	if len(to) == 0 {
		return fmt.Errorf("no recipients specified")
	}

	// Prepare message
	message := es.buildMessage(to, subject, body)

	// Connect to SMTP server
	auth := smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.Host)
	addr := fmt.Sprintf("%s:%d", es.config.Host, es.config.Port)

	var err error
	if es.config.UseTLS {
		err = es.sendWithTLS(addr, auth, es.config.From, to, message)
	} else {
		err = smtp.SendMail(addr, auth, es.config.From, to, []byte(message))
	}

	return err
}

// sendWithTLS sends email using TLS connection
func (es *EmailService) sendWithTLS(addr string, auth smtp.Auth, from string, to []string, message string) error {
	// Connect to SMTP server
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer client.Close()

	// Start TLS if configured
	if es.config.UseTLS {
		if err = client.StartTLS(&tls.Config{ServerName: es.config.Host}); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	// Authenticate
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send email data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	defer w.Close()

	_, err = w.Write([]byte(message))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// buildMessage constructs the email message with proper headers
func (es *EmailService) buildMessage(to []string, subject, body string) string {
	var msg strings.Builder

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", es.config.From))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	msg.WriteString("\r\n")

	// Body
	msg.WriteString(body)

	return msg.String()
}

// SendFormSubmission sends a form submission email
func (es *EmailService) SendFormSubmission(to []string, formData map[string]string) error {
	subject := "New Form Submission"

	var body strings.Builder
	body.WriteString("You have received a new form submission:\n\n")

	for key, value := range formData {
		body.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}

	body.WriteString("\n---\n")
	body.WriteString("This email was sent automatically by staticSend")

	return es.Send(to, subject, body.String())
}

// TestConnection tests the SMTP connection and authentication
func (es *EmailService) TestConnection() error {
	client, err := smtp.Dial(fmt.Sprintf("%s:%d", es.config.Host, es.config.Port))
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if es.config.UseTLS {
		if err := client.StartTLS(&tls.Config{ServerName: es.config.Host}); err != nil {
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	auth := smtp.PlainAuth("", es.config.Username, es.config.Password, es.config.Host)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	return nil
}
