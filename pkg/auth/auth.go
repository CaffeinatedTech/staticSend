package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"staticsend/pkg/models"
)

const (
	// Default cost for bcrypt hashing
	bcryptCost = 12
	// JWT token expiration time
	tokenExpiration = 24 * time.Hour
)

var (
	// ErrInvalidCredentials is returned when email or password is invalid
	ErrInvalidCredentials = errors.New("invalid email or password")
	// ErrUserExists is returned when trying to create a user that already exists
	ErrUserExists = errors.New("user already exists")
	// ErrTokenInvalid is returned when JWT token is invalid
	ErrTokenInvalid = errors.New("invalid token")
)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// CheckPassword compares a password with a bcrypt hash
func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// GenerateToken creates a JWT token for a user
func GenerateToken(user *models.User, secretKey []byte) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.ID,
		"email": user.Email,
		"exp": time.Now().Add(tokenExpiration).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, secretKey []byte) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, ErrTokenInvalid
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrTokenInvalid
}

// GenerateSecretKey generates a random secret key for JWT signing
func GenerateSecretKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate secret key: %w", err)
	}
	return key, nil
}

// GetTokenFromRequest extracts JWT token from Authorization header
func GetTokenFromRequest(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("authorization header format must be: Bearer <token>")
	}

	return parts[1], nil
}

// GetUserIDFromToken extracts user ID from JWT claims
func GetUserIDFromToken(claims jwt.MapClaims) (int64, error) {
	userID, ok := claims["sub"].(float64)
	if !ok {
		return 0, errors.New("invalid user ID in token")
	}
	return int64(userID), nil
}