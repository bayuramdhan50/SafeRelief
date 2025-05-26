package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "userID"

type AuthMiddleware struct {
	jwtSecret []byte
}

func NewAuthMiddleware(jwtSecret []byte) *AuthMiddleware {
	return &AuthMiddleware{jwtSecret: jwtSecret}
}

func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from cookie
		cookie, err := r.Cookie("access_token")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return m.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Extract claims and add user ID to context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), UserIDKey, claims["sub"])
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}
	})
}

type CSRFMiddleware struct {
	secretKey []byte
}

func NewCSRFMiddleware(secretKey []byte) *CSRFMiddleware {
	return &CSRFMiddleware{secretKey: secretKey}
}

func (m *CSRFMiddleware) ValidateCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip CSRF check for GET, HEAD, OPTIONS
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Get CSRF token from header
		token := r.Header.Get("X-CSRF-Token")
		cookie, err := r.Cookie("CSRF-Token")

		if err != nil || cookie == nil || token == "" || token != cookie.Value {
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds comprehensive security headers to all responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Enable XSS protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Strict Transport Security (HTTPS only)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; "+
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
				"style-src 'self' 'unsafe-inline' fonts.googleapis.com; "+
				"font-src 'self' fonts.gstatic.com; "+
				"img-src 'self' data: https:; "+
				"connect-src 'self'; "+
				"frame-ancestors 'none';")

		// Referrer Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions Policy
		w.Header().Set("Permissions-Policy",
			"geolocation=(), microphone=(), camera=(), payment=()")

		next.ServeHTTP(w, r)
	})
}

// RateLimitMiddleware provides IP-based rate limiting
func RateLimitMiddleware(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	limiter := make(map[string]*rateLimitInfo)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractClientIP(r)
			now := time.Now()

			// Clean up old entries
			for k, v := range limiter {
				if now.Sub(v.firstRequest) > window {
					delete(limiter, k)
				}
			}

			info, exists := limiter[ip]
			if !exists {
				limiter[ip] = &rateLimitInfo{
					count:        1,
					firstRequest: now,
				}
				next.ServeHTTP(w, r)
				return
			}

			if now.Sub(info.firstRequest) > window {
				// Reset window
				info.count = 1
				info.firstRequest = now
			} else {
				info.count++
			}

			if info.count > maxRequests {
				w.Header().Set("Retry-After", "3600") // 1 hour
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type rateLimitInfo struct {
	count        int
	firstRequest time.Time
}

// InputSanitizationMiddleware sanitizes request inputs
func InputSanitizationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limit request body size to prevent DoS attacks
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB limit

		next.ServeHTTP(w, r)
	})
}

// extractClientIP extracts the real client IP from request headers
func extractClientIP(r *http.Request) string {
	// Check for real IP in various headers (considering proxy/load balancer)
	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"CF-Connecting-IP", // Cloudflare
		"X-Client-IP",
		"X-Forwarded",
		"Forwarded-For",
		"Forwarded",
	}

	for _, header := range headers {
		if ip := r.Header.Get(header); ip != "" {
			// X-Forwarded-For can contain multiple IPs, take the first one
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					return strings.TrimSpace(ips[0])
				}
			}
			return ip
		}
	}

	// Fallback to RemoteAddr
	if r.RemoteAddr != "" {
		// Remove port if present
		if colonIndex := strings.LastIndex(r.RemoteAddr, ":"); colonIndex != -1 {
			return r.RemoteAddr[:colonIndex]
		}
		return r.RemoteAddr
	}

	return "unknown"
}

func SanitizeInput(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sanitize query parameters
		q := r.URL.Query()
		for key := range q {
			q[key][0] = sanitizeString(q[key][0])
		}
		r.URL.RawQuery = q.Encode()

		// Sanitize path parameters
		r.URL.Path = sanitizeString(r.URL.Path)

		next.ServeHTTP(w, r)
	})
}

func sanitizeString(s string) string {
	// Remove potentially dangerous characters
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, ";", "&#59;")
	return s
}
