package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Context key for session
type contextKey string

const SessionKey contextKey = "session"

// AuthMiddleware creates authentication middleware
func AuthMiddleware(sessionManager *SessionManager, isSecure bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health check
			if r.URL.Path == "/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip auth for GraphQL playground (in development)
			if r.URL.Path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			// Get client IP
			ip := getClientIP(r)
			userAgent := r.UserAgent()

			// Get session cookie
			cookieValue, err := GetSessionCookie(r)
			if err != nil {
				// No session cookie, create new one
				guestID := generateGuestID()
				cookieValue, session, err := sessionManager.CreateSession(guestID, ip, userAgent)
				if err != nil {
					http.Error(w, "Failed to create session", http.StatusInternalServerError)
					return
				}
				SetSessionCookie(w, cookieValue, isSecure)

				// Add session to context
				ctx := context.WithValue(r.Context(), SessionKey, session)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Validate session
			session, err := sessionManager.ValidateSession(cookieValue, ip, userAgent)
			if err != nil {
				// Invalid or expired session, create new one
				guestID := generateGuestID()
				cookieValue, session, err = sessionManager.CreateSession(guestID, ip, userAgent)
				if err != nil {
					http.Error(w, "Failed to create session", http.StatusInternalServerError)
					return
				}
				SetSessionCookie(w, cookieValue, isSecure)

				// Add session to context
				ctx := context.WithValue(r.Context(), SessionKey, session)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// Refresh session expiration
			sessionID := getSessionIDFromCookie(cookieValue)
			sessionManager.RefreshSession(sessionID)

			// Add session to context
			ctx := context.WithValue(r.Context(), SessionKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetSessionFromContext retrieves the session from the context
func GetSessionFromContext(ctx context.Context) *Session {
	if session, ok := ctx.Value(SessionKey).(*Session); ok {
		return session
	}
	return nil
}

func getClientIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	return ip
}

func getSessionIDFromCookie(cookieValue string) string {
	if idx := strings.Index(cookieValue, ":"); idx != -1 {
		return cookieValue[:idx]
	}
	return cookieValue
}

func generateGuestID() string {
	return uuid.New().String()
}
