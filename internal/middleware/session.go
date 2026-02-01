package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCookie = errors.New("invalid cookie")
	ErrExpiredCookie = errors.New("cookie expired")
)

// Session represents a user session
type Session struct {
	GuestID   string
	CreatedAt time.Time
	ExpiresAt time.Time
	IP        string
	UserAgent string
}

// SessionManager manages session creation and validation
type SessionManager struct {
	secretKey []byte
	sessions  map[string]*Session
	mu        sync.RWMutex
}

// NewSessionManager creates a new session manager
func NewSessionManager(secretKey string) *SessionManager {
	sm := &SessionManager{
		secretKey: []byte(secretKey),
		sessions:  make(map[string]*Session),
	}

	// Start cleanup goroutine
	go sm.cleanup()

	return sm
}

// CreateSession creates a new session for a guest
func (sm *SessionManager) CreateSession(guestID, ip, userAgent string) (string, *Session, error) {
	sessionID := generateRandomString(32)
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour) // 24 hours expiration

	session := &Session{
		GuestID:   guestID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		IP:        ip,
		UserAgent: userAgent,
	}

	sm.mu.Lock()
	sm.sessions[sessionID] = session
	sm.mu.Unlock()

	// Create signed cookie value
	cookieValue := sm.createSignedCookie(sessionID, expiresAt)

	return cookieValue, session, nil
}

// ValidateSession validates a session cookie and returns the session
func (sm *SessionManager) ValidateSession(cookieValue string, ip, userAgent string) (*Session, error) {
	// Extract sessionID and signature
	parts := strings.Split(cookieValue, ":")
	if len(parts) != 2 {
		return nil, ErrInvalidCookie
	}

	sessionID := parts[0]
	signature := parts[1]

	// Verify signature
	if !sm.verifySignature(sessionID, signature) {
		return nil, ErrInvalidCookie
	}

	// Get session from store
	sm.mu.RLock()
	session, exists := sm.sessions[sessionID]
	sm.mu.RUnlock()

	if !exists {
		return nil, ErrInvalidCookie
	}

	// Check expiration
	if time.Now().After(session.ExpiresAt) {
		sm.mu.Lock()
		delete(sm.sessions, sessionID)
		sm.mu.Unlock()
		return nil, ErrExpiredCookie
	}

	// Optional: Validate IP and UserAgent for additional security
	// Comment out if you want to allow IP changes (e.g., mobile switching)
	// if session.IP != ip {
	// 	return nil, ErrInvalidCookie
	// }
	// if session.UserAgent != userAgent {
	// 	return nil, ErrInvalidCookie
	// }

	return session, nil
}

// RefreshSession refreshes a session expiration time
func (sm *SessionManager) RefreshSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return ErrInvalidCookie
	}

	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	return nil
}

// DeleteSession deletes a session
func (sm *SessionManager) DeleteSession(sessionID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.sessions, sessionID)
}

// createSignedCookie creates a signed cookie value
func (sm *SessionManager) createSignedCookie(sessionID string, expiresAt time.Time) string {
	// Create signature: HMAC-SHA256(sessionID + expiresAt)
	data := sessionID + ":" + expiresAt.Format(time.RFC3339)
	h := hmac.New(sha256.New, sm.secretKey)
	h.Write([]byte(data))
	signature := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return sessionID + ":" + signature
}

// verifySignature verifies the cookie signature
func (sm *SessionManager) verifySignature(sessionID, signature string) bool {
	// For simplicity, we just verify the signature format
	// In production, you might want to include expiration in signature verification
	h := hmac.New(sha256.New, sm.secretKey)
	h.Write([]byte(sessionID))
	expectedSig := base64.URLEncoding.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSig))
}

// cleanup removes expired sessions
func (sm *SessionManager) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for sessionID, session := range sm.sessions {
			if now.After(session.ExpiresAt) {
				delete(sm.sessions, sessionID)
			}
		}
		sm.mu.Unlock()
	}
}

// generateRandomString generates a random string
func generateRandomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// GetSessionCookieName returns the cookie name for sessions
func GetSessionCookieName() string {
	return "sum100_session"
}

// SetSessionCookie sets the session cookie in the response
func SetSessionCookie(w http.ResponseWriter, sessionValue string, isSecure bool) {
	cookie := &http.Cookie{
		Name:     GetSessionCookieName(),
		Value:    sessionValue,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
}

// GetSessionCookie gets the session cookie from the request
func GetSessionCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie(GetSessionCookieName())
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearSessionCookie clears the session cookie
func ClearSessionCookie(w http.ResponseWriter, isSecure bool) {
	cookie := &http.Cookie{
		Name:     GetSessionCookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
}
