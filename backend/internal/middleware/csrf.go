package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

type CSRFToken struct {
	Token     string
	ExpiresAt time.Time
	Used      bool
}

type CSRFProtection struct {
	tokens    map[string]*CSRFToken
	mutex     sync.RWMutex
	tokenTTL  time.Duration
	secretKey []byte
}

func NewCSRFProtection(secretKey []byte) *CSRFProtection {
	csrf := &CSRFProtection{
		tokens:    make(map[string]*CSRFToken),
		tokenTTL:  15 * time.Minute, // Shorter expiry for better security
		secretKey: secretKey,
	}

	// Start cleanup goroutine
	go csrf.cleanupExpiredTokens()

	return csrf
}

// GenerateToken generates a new CSRF token
func (c *CSRFProtection) GenerateToken() (string, error) {
	// Generate random bytes
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %v", err)
	}

	// Encode to base64
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Store token with expiration
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.tokens[token] = &CSRFToken{
		Token:     token,
		ExpiresAt: time.Now().Add(c.tokenTTL),
		Used:      false,
	}

	return token, nil
}

// ValidateToken validates and consumes a CSRF token
func (c *CSRFProtection) ValidateToken(token string) bool {
	if token == "" {
		return false
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	storedToken, exists := c.tokens[token]
	if !exists {
		return false
	}

	// Check if token is expired
	if time.Now().After(storedToken.ExpiresAt) {
		delete(c.tokens, token)
		return false
	}

	// Check if token was already used (prevent replay attacks)
	if storedToken.Used {
		return false
	}

	// Mark token as used
	storedToken.Used = true

	return true
}

// CSRFMiddleware provides CSRF protection for HTTP requests
func (c *CSRFProtection) CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for GET, HEAD, OPTIONS requests (read-only operations)
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip CSRF check for API endpoints that use other authentication methods
		if strings.HasPrefix(r.URL.Path, "/api/auth/refresh") {
			next.ServeHTTP(w, r)
			return
		}

		// Extract CSRF token from header or form
		token := c.extractToken(r)

		// Validate token
		if !c.ValidateToken(token) {
			http.Error(w, "Invalid or missing CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// extractToken extracts CSRF token from request
func (c *CSRFProtection) extractToken(r *http.Request) string {
	// Try to get token from header first
	if token := r.Header.Get("X-CSRF-Token"); token != "" {
		return token
	}

	// Try to get token from form data
	if err := r.ParseForm(); err == nil {
		if token := r.FormValue("csrf_token"); token != "" {
			return token
		}
	}

	return ""
}

// GetTokenHandler returns a new CSRF token
func (c *CSRFProtection) GetTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := c.GenerateToken()
	if err != nil {
		http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"csrf_token": "%s"}`, token)
}

// cleanupExpiredTokens periodically removes expired tokens
func (c *CSRFProtection) cleanupExpiredTokens() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for token, csrfToken := range c.tokens {
			if now.After(csrfToken.ExpiresAt) || csrfToken.Used {
				delete(c.tokens, token)
			}
		}
		c.mutex.Unlock()
	}
}

// SameSiteMiddleware adds SameSite protection headers
func SameSiteMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add SameSite protection headers
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")

		next.ServeHTTP(w, r)
	})
}

// DoubleSubmitCookie provides additional CSRF protection using double submit cookie pattern
type DoubleSubmitCookie struct {
	secretKey []byte
}

func NewDoubleSubmitCookie(secretKey []byte) *DoubleSubmitCookie {
	return &DoubleSubmitCookie{secretKey: secretKey}
}

func (d *DoubleSubmitCookie) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip for safe methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			// Set CSRF cookie for safe methods
			d.setCSRFCookie(w, r)
			next.ServeHTTP(w, r)
			return
		}

		// Validate CSRF token for unsafe methods
		if !d.validateCSRFToken(r) {
			http.Error(w, "CSRF token validation failed", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (d *DoubleSubmitCookie) setCSRFCookie(w http.ResponseWriter, r *http.Request) {
	// Generate token
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	// Set cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf_token",
		Value:    token,
		HttpOnly: false, // Must be accessible to JavaScript
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   1800, // 30 minutes
	})
}

func (d *DoubleSubmitCookie) validateCSRFToken(r *http.Request) bool {
	// Get token from cookie
	cookie, err := r.Cookie("csrf_token")
	if err != nil {
		return false
	}

	// Get token from header
	headerToken := r.Header.Get("X-CSRF-Token")
	if headerToken == "" {
		return false
	}

	// Compare tokens using constant time comparison
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(headerToken)) == 1
}
