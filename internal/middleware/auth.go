package middleware

import (
	"context"
	"net/http"
	"strings"
)

// Key type for context values
type contextKey string

// Context keys
const (
	UserIDKey contextKey = "user_id"
)

// AuthMiddleware is a middleware for authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")

		// Check if the Authorization header is present and starts with "Bearer "
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract the token
		token := strings.TrimPrefix(authHeader, "Bearer ")

		// In a real application, you would validate the token here
		// For example, you would verify the JWT signature and extract claims

		// For demonstration purposes, we'll just set a dummy user ID in the context
		// In a real application, you would extract the user ID from the token claims
		userID := uint(1)

		// Add the user ID to the request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID gets the user ID from the request context
func GetUserID(r *http.Request) (uint, bool) {
	userID, ok := r.Context().Value(UserIDKey).(uint)
	return userID, ok
}

// RequireAuthentication is a middleware that requires authentication
func RequireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user ID from the context
		_, ok := GetUserID(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
