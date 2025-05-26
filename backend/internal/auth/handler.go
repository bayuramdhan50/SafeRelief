package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp"
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
	validator     *Validator
	auditLogger   *AuditLogger
}

func NewAuthHandler(jwtSecret, refreshSecret []byte, db *sql.DB) *AuthHandler {
	return &AuthHandler{
		jwtSecret:     jwtSecret,
		refreshSecret: refreshSecret,
		db:            db,
		rateLimiter:   NewRateLimiter(100, time.Hour), // 100 requests per hour
		validator:     NewValidator(),
		auditLogger:   NewAuditLogger(db),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Rate limiting check
	ip := r.RemoteAddr
	if !h.rateLimiter.Allow(ip) {
		h.auditLogger.LogSecurityEvent(r, RateLimitExceeded, map[string]string{
			"reason": "Too many login attempts",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Too many login attempts",
		})
		return
	}
	// Check for suspicious activity
	if h.auditLogger.DetectSuspiciousActivity(ip, 30*time.Minute) {
		h.auditLogger.LogSecurityEvent(r, SuspiciousActivity, map[string]string{
			"reason": "Multiple failed attempts detected",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Account temporarily restricted due to suspicious activity",
		})
		return
	}
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		h.auditLogger.LogSecurityEvent(r, ValidationFailed, map[string]string{
			"reason": "Invalid request body",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Sanitize and validate input
	creds.Email = h.validator.SanitizeInput(creds.Email)
	creds.Password = h.validator.SanitizeInput(creds.Password)
	creds.MFACode = h.validator.SanitizeInput(creds.MFACode)
	// Validate email format
	if emailErrors := h.validator.ValidateEmail(creds.Email); len(emailErrors) > 0 {
		h.auditLogger.LogAuthEvent(r, ValidationFailed, "", creds.Email, map[string]string{
			"reason": "Invalid email format",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid email format",
		})
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
			// Log failed login attempt
			h.auditLogger.LogAuthEvent(r, LoginFailed, "", creds.Email, map[string]string{
				"reason": "User not found",
			})
			// Use same error message as password mismatch for security
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid credentials",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal server error",
		})
		return
	}
	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		h.auditLogger.LogAuthEvent(r, LoginFailed, user.ID, user.Email, map[string]string{
			"reason":       "Account locked",
			"locked_until": user.LockedUntil.Format(time.RFC3339),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Account is temporarily locked",
		})
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
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			})
			return
		}

		// Log failed login attempt
		details := map[string]string{
			"reason":          "Invalid password",
			"failed_attempts": fmt.Sprintf("%d", newFailedAttempts),
		}
		if lockedUntil != nil {
			details["account_locked"] = "true"
			details["locked_until"] = lockedUntil.Format(time.RFC3339)
			h.auditLogger.LogAuthEvent(r, AccountLocked, user.ID, user.Email, details)
		} else {
			h.auditLogger.LogAuthEvent(r, LoginFailed, user.ID, user.Email, details)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid credentials",
		})
		return
	}
	// Reset failed attempts on successful password verification
	_, err = h.db.Exec(
		"UPDATE users SET failed_attempts = 0, locked_until = NULL WHERE id = ?",
		user.ID,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal server error",
		})
		return
	}
	// Check MFA if enabled
	if user.MFAEnabled {
		if creds.MFACode == "" {
			h.auditLogger.LogAuthEvent(r, LoginFailed, user.ID, user.Email, map[string]string{
				"reason": "MFA code required but not provided",
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "MFA required",
			})
			return
		}
		if !totp.Validate(creds.MFACode, user.MFASecret) {
			h.auditLogger.LogAuthEvent(r, LoginMFAFailed, user.ID, user.Email, map[string]string{
				"reason": "Invalid MFA code",
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid MFA code",
			})
			return
		}
	}
	// Generate tokens
	accessToken, err := h.generateAccessToken(user.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error generating access token",
		})
		return
	}

	refreshToken, err := h.generateRefreshToken(user.ID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error generating refresh token",
		})
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

	// Log successful login
	h.auditLogger.LogAuthEvent(r, LoginSuccess, user.ID, user.Email, map[string]string{
		"login_method": func() string {
			if user.MFAEnabled {
				return "password_and_mfa"
			}
			return "password_only"
		}(),
	})
	// Return user data (excluding sensitive information)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": map[string]interface{}{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Rate limiting check
	ip := r.RemoteAddr
	if !h.rateLimiter.Allow(ip) {
		h.auditLogger.LogSecurityEvent(r, RateLimitExceeded, map[string]string{
			"reason": "Too many registration attempts",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Too many registration attempts",
		})
		return
	}

	var user struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		h.auditLogger.LogSecurityEvent(r, ValidationFailed, map[string]string{
			"reason": "Invalid request body",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	// Sanitize inputs
	user.Username = h.validator.SanitizeInput(user.Username)
	user.Email = h.validator.SanitizeInput(user.Email)
	user.Password = h.validator.SanitizeInput(user.Password)

	// Collect all validation errors
	var allErrors []ValidationError

	// Validate username
	if usernameErrors := h.validator.ValidateUsername(user.Username); len(usernameErrors) > 0 {
		allErrors = append(allErrors, usernameErrors...)
	}

	// Validate email
	if emailErrors := h.validator.ValidateEmail(user.Email); len(emailErrors) > 0 {
		allErrors = append(allErrors, emailErrors...)
	}

	// Validate password
	if passwordErrors := h.validator.ValidatePassword(user.Password); len(passwordErrors) > 0 {
		allErrors = append(allErrors, passwordErrors...)
	}

	// If there are validation errors, return them
	if len(allErrors) > 0 {
		h.auditLogger.LogAuthEvent(r, ValidationFailed, "", user.Email, map[string]string{
			"reason": "Input validation failed",
			"errors": fmt.Sprintf("%d validation errors", len(allErrors)),
		})

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": allErrors,
		})
		return
	}
	// Check if email already exists
	var existingUserID string
	err := h.db.QueryRow("SELECT id FROM users WHERE email = ?", user.Email).Scan(&existingUserID)
	if err == nil {
		h.auditLogger.LogAuthEvent(r, RegisterFailed, "", user.Email, map[string]string{
			"reason": "Email already registered",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Email already registered",
		})
		return
	} else if err != sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal server error",
		})
		return
	}
	// Check if username already exists
	err = h.db.QueryRow("SELECT id FROM users WHERE username = ?", user.Username).Scan(&existingUserID)
	if err == nil {
		h.auditLogger.LogAuthEvent(r, RegisterFailed, "", user.Email, map[string]string{
			"reason": "Username already taken",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Username already taken",
		})
		return
	} else if err != sql.ErrNoRows {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Internal server error",
		})
		return
	}

	// Generate password hash with higher cost for better security
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error hashing password"})
		return
	}

	// Generate MFA secret
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SafeRelief",
		AccountName: user.Email,
		Period:      30,
		Digits:      6,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error generating MFA secret",
		})
		return
	}
	// Insert user into database
	result, err := h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, mfa_secret, created_at, updated_at)
		VALUES (UUID_TO_BIN(UUID()), ?, ?, ?, ?, NOW(), NOW())`,
		user.Username, user.Email, hashedPassword, secret.Secret(),
	)
	if err != nil {
		// Check for duplicate email or username
		if mysqlErr, ok := err.(*mysql.MySQLError); ok && mysqlErr.Number == 1062 {
			h.auditLogger.LogAuthEvent(r, RegisterFailed, "", user.Email, map[string]string{
				"reason": "Duplicate key constraint",
			})
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Email or username already exists",
			})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error creating user",
		})
		return
	}
	// Get the created user ID
	userID, err := result.LastInsertId()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Error retrieving user ID"})
		return
	}

	// Log successful registration
	h.auditLogger.LogAuthEvent(r, RegisterSuccess, fmt.Sprintf("%d", userID), user.Email, map[string]string{
		"username": user.Username,
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "User registered successfully",
		"user": map[string]interface{}{
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from token if available for logging
	userID := ""
	if cookie, err := r.Cookie("access_token"); err == nil {
		if token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
			return h.jwtSecret, nil
		}); err == nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				if sub, exists := claims["sub"].(string); exists {
					userID = sub
				}
			}
		}
	}

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

	// Log logout event
	h.auditLogger.LogAuthEvent(r, LogoutSuccess, userID, "", map[string]string{
		"logout_method": "manual",
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

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// Get refresh token from cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "Refresh token not found", http.StatusUnauthorized)
		return
	}

	// Parse and validate refresh token
	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.refreshSecret, nil
	})

	if err != nil || !token.Valid {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
		return
	}

	// Verify user still exists and is not locked
	var user User
	err = h.db.QueryRow(`
		SELECT id, username, email, mfa_enabled, failed_attempts, locked_until 
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.MFAEnabled, &user.FailedAttempts, &user.LockedUntil)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		http.Error(w, "Account is temporarily locked", http.StatusUnauthorized)
		return
	}

	// Generate new access token
	accessToken, err := h.generateAccessToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	// Generate new refresh token
	newRefreshToken, err := h.generateRefreshToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Set new refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	// Return new access token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Token refreshed successfully",
		"accessToken": accessToken,
		"user": map[string]interface{}{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"mfaEnabled": user.MFAEnabled,
		},
	})
}

// ChangePassword handles password change requests with enhanced security
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from authenticated request
	userID := r.Context().Value("userID").(string)

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
		MFACode         string `json:"mfaCode,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.auditLogger.LogSecurityEvent(r, ValidationFailed, map[string]string{
			"reason": "Invalid request body for password change",
		})
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Sanitize inputs
	req.CurrentPassword = h.validator.SanitizeInput(req.CurrentPassword)
	req.NewPassword = h.validator.SanitizeInput(req.NewPassword)
	req.MFACode = h.validator.SanitizeInput(req.MFACode)

	// Validate new password
	if passwordErrors := h.validator.ValidatePassword(req.NewPassword); len(passwordErrors) > 0 {
		h.auditLogger.LogAuthEvent(r, ValidationFailed, userID, "", map[string]string{
			"reason": "New password validation failed",
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": passwordErrors,
		})
		return
	}

	// Get user from database
	var user User
	err := h.db.QueryRow(
		"SELECT id, username, email, password_hash, mfa_secret, mfa_enabled FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.MFASecret, &user.MFAEnabled)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		h.auditLogger.LogAuthEvent(r, ValidationFailed, userID, user.Email, map[string]string{
			"reason": "Current password verification failed",
		})
		http.Error(w, "Current password is incorrect", http.StatusUnauthorized)
		return
	}

	// Check if new password is same as current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.NewPassword)); err == nil {
		h.auditLogger.LogAuthEvent(r, ValidationFailed, userID, user.Email, map[string]string{
			"reason": "New password same as current password",
		})
		http.Error(w, "New password must be different from current password", http.StatusBadRequest)
		return
	}

	// Verify MFA if enabled
	if user.MFAEnabled {
		if req.MFACode == "" {
			http.Error(w, "MFA code required for password change", http.StatusUnauthorized)
			return
		}

		if !totp.Validate(req.MFACode, user.MFASecret) {
			h.auditLogger.LogAuthEvent(r, ValidationFailed, userID, user.Email, map[string]string{
				"reason": "MFA verification failed for password change",
			})
			http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
			return
		}
	}

	// Hash new password
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), 12)
	if err != nil {
		http.Error(w, "Error hashing new password", http.StatusInternalServerError)
		return
	}

	// Update password in database
	_, err = h.db.Exec(
		"UPDATE users SET password_hash = ?, updated_at = NOW() WHERE id = ?",
		newHashedPassword, userID,
	)
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	// Log password change
	h.auditLogger.LogAuthEvent(r, PasswordChanged, userID, user.Email, map[string]string{
		"change_method": func() string {
			if user.MFAEnabled {
				return "password_and_mfa"
			}
			return "password_only"
		}(),
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password changed successfully",
	})
}
