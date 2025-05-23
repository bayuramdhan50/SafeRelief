package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	MFACode  string `json:"mfaCode,omitempty"`
}

type User struct {
	ID             string     `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	PasswordHash   string     `json:"-"`
	MFASecret      string     `json:"-"`
	MFAEnabled     bool       `json:"mfaEnabled"`
	FailedAttempts int        `json:"-"`
	LockedUntil    *time.Time `json:"-"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

type AuthHandler struct {
	jwtSecret     []byte
	refreshSecret []byte
	db            *sql.DB
	rateLimiter   *RateLimiter
}

func NewAuthHandler(jwtSecret, refreshSecret []byte, db *sql.DB) *AuthHandler {
	return &AuthHandler{
		jwtSecret:     jwtSecret,
		refreshSecret: refreshSecret,
		db:            db,
		rateLimiter:   NewRateLimiter(100, time.Hour), // 100 requests per hour
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Rate limiting check
	ip := r.RemoteAddr
	if !h.rateLimiter.Allow(ip) {
		http.Error(w, "Too many login attempts", http.StatusTooManyRequests)
		return
	}

	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from database
	var user User
	err := h.db.QueryRow(
		"SELECT id, username, email, password_hash, mfa_secret, mfa_enabled, failed_attempts, locked_until FROM users WHERE email = ?",
		creds.Email,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.MFASecret, &user.MFAEnabled, &user.FailedAttempts, &user.LockedUntil)

	if err != nil {
		if err == sql.ErrNoRows {
			// Use same error message as password mismatch for security
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		http.Error(w, "Account is temporarily locked", http.StatusForbidden)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		// Increment failed attempts
		newFailedAttempts := user.FailedAttempts + 1
		var lockedUntil *time.Time

		if newFailedAttempts >= 5 {
			t := time.Now().Add(15 * time.Minute)
			lockedUntil = &t
		}

		_, err := h.db.Exec(
			"UPDATE users SET failed_attempts = ?, locked_until = ? WHERE id = ?",
			newFailedAttempts, lockedUntil, user.ID,
		)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Reset failed attempts on successful password verification
	_, err = h.db.Exec(
		"UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE id = ?",
		user.ID,
	)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check MFA if enabled
	if user.MFAEnabled {
		if creds.MFACode == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "MFA required",
			})
			return
		}

		if !totp.Validate(creds.MFACode, user.MFASecret) {
			http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
			return
		}
	}

	// Generate tokens
	accessToken, err := h.generateAccessToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.generateRefreshToken(user.ID)
	if err != nil {
		http.Error(w, "Error generating refresh token", http.StatusInternalServerError)
		return
	}

	// Set tokens in secure HTTP-only cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   900, // 15 minutes
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   604800, // 7 days
	})

	// Return user data (excluding sensitive information)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate password hash
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Generate MFA secret
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SafeRelief",
		AccountName: user.Email,
	})
	if err != nil {
		http.Error(w, "Error generating MFA secret", http.StatusInternalServerError)
		return
	}

	// Insert user into database
	result, err := h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, mfa_secret, created_at, updated_at)
		VALUES (UUID_TO_BIN(UUID()), ?, ?, ?, ?, NOW(), NOW())`,
		user.Username, user.Email, hashedPassword, secret.Secret(),
	)
	if err != nil {
		// Check for duplicate email
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear cookies
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) generateAccessToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(15 * time.Minute).Unix(),
	})

	return token.SignedString(h.jwtSecret)
}

func (h *AuthHandler) generateRefreshToken(userID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})

	return token.SignedString(h.refreshSecret)
}

type RateLimiter struct {
	requests map[string]*requestCount
	limit    int
	window   time.Duration
	mu       sync.Mutex
}

type requestCount struct {
	count     int
	startTime time.Time
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string]*requestCount),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if req, exists := rl.requests[key]; exists {
		if now.Sub(req.startTime) > rl.window {
			rl.requests[key] = &requestCount{1, now}
			return true
		}

		if req.count >= rl.limit {
			return false
		}

		req.count++
		return true
	}

	rl.requests[key] = &requestCount{1, now}
	return true
}
