package middleware

import (
	"context"
	"net/http"
	"strings"

	"staticsend/pkg/auth"
	"staticsend/pkg/database"
	"staticsend/pkg/models"
)

// Context keys for storing authentication data
type contextKey string

const (
	// UserKey is the context key for storing user object
	UserKey contextKey = "user"
	// ClaimsKey is the context key for storing JWT claims
	ClaimsKey contextKey = "claims"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	SecretKey []byte
	DB        *database.Database
	// Optional: paths that don't require authentication
	PublicPaths []string
}

// AuthMiddleware provides JWT authentication middleware with cookie support
func AuthMiddleware(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for public paths
			if isPublicPath(r.URL.Path, config.PublicPaths) {
				next.ServeHTTP(w, r)
				return
			}

			var tokenString string
			var err error

			// First try to get token from Authorization header
			tokenString, err = auth.GetTokenFromRequest(r)
			if err != nil {
				// If no Authorization header, try to get from cookie
				if cookie, err := r.Cookie("auth_token"); err == nil {
					tokenString = cookie.Value
				} else {
					// For web requests, redirect to login instead of 401 error
					if r.Header.Get("HX-Request") == "true" {
						// HTMX request, return 401
						http.Error(w, "Unauthorized: authentication required", http.StatusUnauthorized)
					} else {
						// Regular browser request, redirect to login
						http.Redirect(w, r, "/login", http.StatusFound)
					}
					return
				}
			}

			claims, err := auth.ValidateToken(tokenString, config.SecretKey)
			if err != nil {
				// Invalid token - clear the bad cookie and redirect to login
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    "",
					Path:     "/",
					HttpOnly: true,
					Secure:   false,
					MaxAge:   -1,
				})
				
				if r.Header.Get("HX-Request") == "true" {
					http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				} else {
					http.Redirect(w, r, "/login", http.StatusFound)
				}
				return
			}

			userID, err := auth.GetUserIDFromToken(claims)
			if err != nil {
				http.Error(w, "Unauthorized: invalid token claims", http.StatusUnauthorized)
				return
			}

			// Get user from database
			user, err := models.GetUserByID(config.DB.Connection, userID)
			if err != nil || user == nil {
				// User not found - clear the bad cookie and redirect to login
				http.SetCookie(w, &http.Cookie{
					Name:     "auth_token",
					Value:    "",
					Path:     "/",
					HttpOnly: true,
					Secure:   false,
					MaxAge:   -1,
				})
				
				if r.Header.Get("HX-Request") == "true" {
					http.Error(w, "Unauthorized: user not found", http.StatusUnauthorized)
				} else {
					http.Redirect(w, r, "/login", http.StatusFound)
				}
				return
			}

			// Add user and claims to context
			ctx := context.WithValue(r.Context(), UserKey, user)
			ctx = context.WithValue(ctx, ClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext retrieves the user from request context
func GetUserFromContext(ctx context.Context) (*models.User, bool) {
	user, ok := ctx.Value(UserKey).(*models.User)
	return user, ok
}

// GetClaimsFromContext retrieves the JWT claims from request context
func GetClaimsFromContext(ctx context.Context) (map[string]interface{}, bool) {
	claims, ok := ctx.Value(ClaimsKey).(map[string]interface{})
	return claims, ok
}

// isPublicPath checks if the current path should bypass authentication
func isPublicPath(path string, publicPaths []string) bool {
	for _, publicPath := range publicPaths {
		if path == publicPath || strings.HasPrefix(path, publicPath+"/") {
			return true
		}
	}
	return false
}