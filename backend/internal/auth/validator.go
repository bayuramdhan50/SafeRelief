package auth

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// ValidateEmail validates email format and checks for common issues
func (v *Validator) ValidateEmail(email string) []ValidationError {
	var errors []ValidationError

	if email == "" {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Email is required",
		})
		return errors
	}

	// Trim whitespace
	email = strings.TrimSpace(email)

	// Basic length check
	if len(email) > 254 {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Email is too long (maximum 254 characters)",
		})
	}

	// Email format validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		errors = append(errors, ValidationError{
			Field:   "email",
			Message: "Invalid email format",
		})
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
	}

	emailLower := strings.ToLower(email)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(emailLower, pattern) {
			errors = append(errors, ValidationError{
				Field:   "email",
				Message: "Email contains invalid characters",
			})
			break
		}
	}

	return errors
}

// ValidatePassword validates password strength and security
func (v *Validator) ValidatePassword(password string) []ValidationError {
	var errors []ValidationError

	if password == "" {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password is required",
		})
		return errors
	}

	// Length requirements
	if len(password) < 8 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must be at least 8 characters long",
		})
	}

	if len(password) > 128 {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password is too long (maximum 128 characters)",
		})
	}

	// Character type requirements
	var hasUpper, hasLower, hasDigit, hasSpecial bool
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must contain at least one uppercase letter",
		})
	}

	if !hasLower {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must contain at least one lowercase letter",
		})
	}

	if !hasDigit {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must contain at least one digit",
		})
	}

	if !hasSpecial {
		errors = append(errors, ValidationError{
			Field:   "password",
			Message: "Password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)",
		})
	}

	// Check for common weak patterns
	weakPatterns := []string{
		"123456",
		"password",
		"qwerty",
		"abc123",
		"admin",
		"letmein",
		"welcome",
		"monkey",
		"dragon",
		"master",
	}

	passwordLower := strings.ToLower(password)
	for _, pattern := range weakPatterns {
		if strings.Contains(passwordLower, pattern) {
			errors = append(errors, ValidationError{
				Field:   "password",
				Message: "Password contains common weak patterns",
			})
			break
		}
	}

	// Check for repeated characters
	repeatedCount := 0
	var prevChar rune
	for _, char := range password {
		if char == prevChar {
			repeatedCount++
			if repeatedCount >= 3 {
				errors = append(errors, ValidationError{
					Field:   "password",
					Message: "Password cannot contain more than 3 consecutive identical characters",
				})
				break
			}
		} else {
			repeatedCount = 0
		}
		prevChar = char
	}

	return errors
}

// ValidateUsername validates username format and security
func (v *Validator) ValidateUsername(username string) []ValidationError {
	var errors []ValidationError

	if username == "" {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username is required",
		})
		return errors
	}

	// Trim whitespace
	username = strings.TrimSpace(username)

	// Length requirements
	if len(username) < 3 {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username must be at least 3 characters long",
		})
	}

	if len(username) > 50 {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username is too long (maximum 50 characters)",
		})
	}

	// Character validation
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		errors = append(errors, ValidationError{
			Field:   "username",
			Message: "Username can only contain letters, numbers, underscores, and hyphens",
		})
	}

	// Check for reserved usernames
	reservedNames := []string{
		"admin", "administrator", "root", "system", "user", "guest",
		"test", "demo", "api", "www", "mail", "ftp", "support",
		"info", "help", "contact", "sales", "marketing", "security",
		"null", "undefined", "anonymous", "public", "private",
	}

	usernameLower := strings.ToLower(username)
	for _, reserved := range reservedNames {
		if usernameLower == reserved {
			errors = append(errors, ValidationError{
				Field:   "username",
				Message: "This username is reserved and cannot be used",
			})
			break
		}
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"<script",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
		"eval(",
		"alert(",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(usernameLower, pattern) {
			errors = append(errors, ValidationError{
				Field:   "username",
				Message: "Username contains invalid characters",
			})
			break
		}
	}

	return errors
}

// SanitizeInput removes potentially dangerous characters
func (v *Validator) SanitizeInput(input string) string {
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")

	// Trim whitespace
	input = strings.TrimSpace(input)

	// Remove control characters except newlines and tabs
	result := make([]rune, 0, len(input))
	for _, r := range input {
		if r >= 32 || r == '\n' || r == '\t' {
			result = append(result, r)
		}
	}

	return string(result)
}

// ValidateGenericText validates general text input
func (v *Validator) ValidateGenericText(text string, fieldName string, minLength, maxLength int) []ValidationError {
	var errors []ValidationError

	if text == "" && minLength > 0 {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fieldName + " is required",
		})
		return errors
	}

	// Trim whitespace
	text = strings.TrimSpace(text)

	// Length validation
	if len(text) < minLength {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s must be at least %d characters long", fieldName, minLength),
		})
	}

	if len(text) > maxLength {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%s is too long (maximum %d characters)", fieldName, maxLength),
		})
	}

	// Check for suspicious patterns that might indicate XSS or injection attacks
	suspiciousPatterns := []string{
		"<script",
		"</script>",
		"javascript:",
		"vbscript:",
		"onload=",
		"onerror=",
		"onclick=",
		"onmouseover=",
		"eval(",
		"expression(",
		"alert(",
		"confirm(",
		"prompt(",
		"document.cookie",
		"document.write",
		"window.location",
		"<iframe",
		"<object",
		"<embed",
		"<link",
		"<meta",
		"<style",
		"url(",
		"@import",
	}

	textLower := strings.ToLower(text)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(textLower, pattern) {
			errors = append(errors, ValidationError{
				Field:   fieldName,
				Message: fieldName + " contains potentially unsafe content",
			})
			break
		}
	}

	return errors
}
