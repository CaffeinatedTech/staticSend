package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"staticsend/pkg/models"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("HashPassword returned empty string")
	}

	// Test that the same password produces different hashes (due to salt)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed on second call: %v", err)
	}

	if hash == hash2 {
		t.Error("HashPassword produced identical hashes for same password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Test correct password
	err = CheckPassword(password, hash)
	if err != nil {
		t.Errorf("CheckPassword failed with correct password: %v", err)
	}

	// Test incorrect password
	err = CheckPassword("wrongpassword", hash)
	if err == nil {
		t.Error("CheckPassword should have failed with incorrect password")
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	user := &models.User{
		ID:    1,
		Email: "test@example.com",
	}

	secretKey, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("GenerateSecretKey failed: %v", err)
	}

	// Generate token
	tokenString, err := GenerateToken(user, secretKey)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if tokenString == "" {
		t.Error("GenerateToken returned empty string")
	}

	// Validate token
	claims, err := ValidateToken(tokenString, secretKey)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Check claims
	if claims["sub"] != float64(user.ID) {
		t.Errorf("Expected sub claim %d, got %v", user.ID, claims["sub"])
	}

	if claims["email"] != user.Email {
		t.Errorf("Expected email claim %s, got %v", user.Email, claims["email"])
	}

	// Test with wrong secret key
	wrongKey := []byte("wrong-secret-key")
	_, err = ValidateToken(tokenString, wrongKey)
	if err == nil {
		t.Error("ValidateToken should have failed with wrong secret key")
	}

	// Test with expired token (manipulate claims)
	expiredClaims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
		"iat":   time.Now().Add(-2 * time.Hour).Unix(),
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, err := expiredToken.SignedString(secretKey)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	_, err = ValidateToken(expiredTokenString, secretKey)
	if err == nil {
		t.Error("ValidateToken should have failed with expired token")
	}
}

func TestGenerateSecretKey(t *testing.T) {
	key1, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("GenerateSecretKey failed: %v", err)
	}

	key2, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("GenerateSecretKey failed on second call: %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("Expected key length 32, got %d", len(key1))
	}

	if string(key1) == string(key2) {
		t.Error("GenerateSecretKey produced identical keys")
	}
}