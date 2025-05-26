package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
)

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

type UserHandler struct {
	db *sql.DB
}

func NewUserHandler(db *sql.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var user User
	err := h.db.QueryRow(`
		SELECT id, username, email, mfa_enabled, created_at, updated_at 
		FROM users WHERE id = ?
	`, userID).Scan(&user.ID, &user.Username, &user.Email, &user.MFAEnabled, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var updateData struct {
		Username string `json:"username"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err := h.db.Exec(`
		UPDATE users SET username = ?, email = ?, updated_at = NOW() 
		WHERE id = ?
	`, updateData.Username, updateData.Email, userID)

	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Profile updated successfully"})
}

func (h *UserHandler) EnableMFA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Generate TOTP secret
	secret, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "SafeRelief",
		AccountName: userID,
	})
	if err != nil {
		http.Error(w, "Failed to generate MFA secret", http.StatusInternalServerError)
		return
	}

	// Save secret to database
	_, err = h.db.Exec(`
		UPDATE users SET mfa_secret = ?, mfa_enabled = true, updated_at = NOW() 
		WHERE id = ?
	`, secret.Secret(), userID)

	if err != nil {
		http.Error(w, "Failed to enable MFA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "MFA enabled successfully",
		"qrCode":  secret.URL(),
		"secret":  secret.Secret(),
	})
}

func (h *UserHandler) DisableMFA(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	var requestData struct {
		Password string `json:"password"`
		MFACode  string `json:"mfaCode"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Verify password and MFA code before disabling
	var passwordHash, mfaSecret string
	err := h.db.QueryRow(`
		SELECT password_hash, mfa_secret FROM users WHERE id = ?
	`, userID).Scan(&passwordHash, &mfaSecret)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(requestData.Password)); err != nil {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	// Verify MFA code
	if !totp.Validate(requestData.MFACode, mfaSecret) {
		http.Error(w, "Invalid MFA code", http.StatusUnauthorized)
		return
	}

	// Disable MFA
	_, err = h.db.Exec(`
		UPDATE users SET mfa_secret = '', mfa_enabled = false, updated_at = NOW() 
		WHERE id = ?
	`, userID)

	if err != nil {
		http.Error(w, "Failed to disable MFA", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "MFA disabled successfully"})
}
