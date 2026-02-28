package middleware

import (
	"net/http"
	"strings"

	"github.com/kenshero/100sumgame/internal/config"
)

// AdminAuthMiddleware validates admin secret token from Authorization header
// This middleware should be applied to GraphQL endpoint for admin-only mutations
func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		// Expected format: "Bearer <token>"
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			http.Error(w, "Invalid authorization header format. Expected: 'Bearer <token>'", http.StatusUnauthorized)
			return
		}

		// Verify token against hardcoded admin secret
		if token != config.AdminSecretToken {
			http.Error(w, "Invalid admin token", http.StatusUnauthorized)
			return
		}

		// Token is valid, proceed to next handler
		next.ServeHTTP(w, r)
	})
}
