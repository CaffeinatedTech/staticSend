package turnstile

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultVerifyURL is the Cloudflare Turnstile verification endpoint
	DefaultVerifyURL = "https://challenges.cloudflare.com/turnstile/v0/siteverify"

	// DefaultTimeout is the default timeout for verification requests
	DefaultTimeout = 10 * time.Second
)

// VerificationResponse represents the response from Cloudflare Turnstile verification
type VerificationResponse struct {
	Success     bool     `json:"success"`
	ChallengeTS string   `json:"challenge_ts,omitempty"`
	Hostname    string   `json:"hostname,omitempty"`
	ErrorCodes  []string `json:"error-codes,omitempty"`
	Action      string   `json:"action,omitempty"`
	CData       string   `json:"cdata,omitempty"`
}

// Validator handles Cloudflare Turnstile token validation
type Validator struct {
	secretKey  string
	verifyURL  string
	httpClient *http.Client
}

// NewValidator creates a new Turnstile validator
func NewValidator(secretKey string) *Validator {
	return &Validator{
		secretKey: secretKey,
		verifyURL: DefaultVerifyURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}
}

// WithVerifyURL sets a custom verification URL (for testing)
func (v *Validator) WithVerifyURL(url string) *Validator {
	v.verifyURL = url
	return v
}

// WithHTTPClient sets a custom HTTP client (for testing)
func (v *Validator) WithHTTPClient(client *http.Client) *Validator {
	v.httpClient = client
	return v
}

// Verify validates a Turnstile token with optional remote IP
func (v *Validator) Verify(ctx context.Context, token, remoteIP string) (*VerificationResponse, error) {
	if token == "" {
		return &VerificationResponse{
			Success:    false,
			ErrorCodes: []string{"missing-input-response"},
		}, nil
	}

	// Prepare form data
	form := url.Values{}
	form.Add("secret", v.secretKey)
	form.Add("response", token)
	if remoteIP != "" {
		form.Add("remoteip", remoteIP)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", v.verifyURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute request
	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("verification request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var verificationResponse VerificationResponse
	if err := json.Unmarshal(body, &verificationResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &verificationResponse, nil
}

// IsValid checks if the verification response indicates a successful validation
func (vr *VerificationResponse) IsValid() bool {
	return vr.Success
}

// HasError checks if the verification response contains specific error codes
func (vr *VerificationResponse) HasError(errorCode string) bool {
	for _, code := range vr.ErrorCodes {
		if code == errorCode {
			return true
		}
	}
	return false
}
