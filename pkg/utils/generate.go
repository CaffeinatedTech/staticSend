package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// GenerateFormKey creates a unique, URL-safe form key
func GenerateFormKey() (string, error) {
	// Generate 18 random bytes (24 base64 characters)
	bytes := make([]byte, 18)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	
	// Encode to base64 URL-safe format and remove padding
	key := base64.URLEncoding.EncodeToString(bytes)
	key = strings.TrimRight(key, "=")
	
	return key, nil
}