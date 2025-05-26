package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type AuditEventType string

const (
	LoginSuccess       AuditEventType = "LOGIN_SUCCESS"
	LoginFailed        AuditEventType = "LOGIN_FAILED"
	LoginMFAFailed     AuditEventType = "LOGIN_MFA_FAILED"
	RegisterSuccess    AuditEventType = "REGISTER_SUCCESS"
	RegisterFailed     AuditEventType = "REGISTER_FAILED"
	PasswordChanged    AuditEventType = "PASSWORD_CHANGED"
	MFAEnabled         AuditEventType = "MFA_ENABLED"
	MFADisabled        AuditEventType = "MFA_DISABLED"
	AccountLocked      AuditEventType = "ACCOUNT_LOCKED"
	AccountUnlocked    AuditEventType = "ACCOUNT_UNLOCKED"
	TokenRefreshed     AuditEventType = "TOKEN_REFRESHED"
	LogoutSuccess      AuditEventType = "LOGOUT_SUCCESS"
	SuspiciousActivity AuditEventType = "SUSPICIOUS_ACTIVITY"
	RateLimitExceeded  AuditEventType = "RATE_LIMIT_EXCEEDED"
	ValidationFailed   AuditEventType = "VALIDATION_FAILED"
)

type AuditEvent struct {
	ID        string            `json:"id"`
	UserID    string            `json:"userId,omitempty"`
	Email     string            `json:"email,omitempty"`
	EventType AuditEventType    `json:"eventType"`
	IPAddress string            `json:"ipAddress"`
	UserAgent string            `json:"userAgent"`
	Details   map[string]string `json:"details,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Severity  string            `json:"severity"`
}

type AuditLogger struct {
	db *sql.DB
}

func NewAuditLogger(db *sql.DB) *AuditLogger {
	return &AuditLogger{db: db}
}

// LogEvent logs an audit event to the database and system logs
func (al *AuditLogger) LogEvent(event AuditEvent) {
	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Set severity based on event type
	if event.Severity == "" {
		event.Severity = al.getSeverityForEventType(event.EventType)
	}

	// Log to database
	go al.logToDatabase(event)

	// Log to system logs for immediate monitoring
	al.logToSystem(event)
}

// LogAuthEvent is a convenience method for logging authentication events
func (al *AuditLogger) LogAuthEvent(r *http.Request, eventType AuditEventType, userID, email string, details map[string]string) {
	event := AuditEvent{
		UserID:    userID,
		Email:     email,
		EventType: eventType,
		IPAddress: al.extractIPAddress(r),
		UserAgent: r.UserAgent(),
		Details:   details,
		Timestamp: time.Now(),
	}

	al.LogEvent(event)
}

// LogSecurityEvent logs security-related events with high severity
func (al *AuditLogger) LogSecurityEvent(r *http.Request, eventType AuditEventType, details map[string]string) {
	event := AuditEvent{
		EventType: eventType,
		IPAddress: al.extractIPAddress(r),
		UserAgent: r.UserAgent(),
		Details:   details,
		Timestamp: time.Now(),
		Severity:  "HIGH",
	}

	al.LogEvent(event)
}

func (al *AuditLogger) logToDatabase(event AuditEvent) {
	if al.db == nil {
		log.Printf("Database not available for audit logging")
		return
	}

	detailsJSON, _ := json.Marshal(event.Details)

	query := `
		INSERT INTO audit_logs (user_id, email, event_type, ip_address, user_agent, details, severity, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := al.db.Exec(query,
		event.UserID,
		event.Email,
		string(event.EventType),
		event.IPAddress,
		event.UserAgent,
		string(detailsJSON),
		event.Severity,
		event.Timestamp,
	)

	if err != nil {
		log.Printf("Failed to log audit event to database: %v", err)
	}
}

func (al *AuditLogger) logToSystem(event AuditEvent) {
	logMessage := map[string]interface{}{
		"audit_event": true,
		"event_type":  event.EventType,
		"user_id":     event.UserID,
		"email":       event.Email,
		"ip_address":  event.IPAddress,
		"user_agent":  event.UserAgent,
		"severity":    event.Severity,
		"timestamp":   event.Timestamp.Format(time.RFC3339),
		"details":     event.Details,
	}

	jsonLog, _ := json.Marshal(logMessage)

	// Use different log levels based on severity
	switch event.Severity {
	case "HIGH", "CRITICAL":
		log.Printf("SECURITY ALERT: %s", string(jsonLog))
	case "MEDIUM":
		log.Printf("SECURITY WARNING: %s", string(jsonLog))
	default:
		log.Printf("AUDIT: %s", string(jsonLog))
	}
}

func (al *AuditLogger) extractIPAddress(r *http.Request) string {
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

func (al *AuditLogger) getSeverityForEventType(eventType AuditEventType) string {
	switch eventType {
	case SuspiciousActivity, AccountLocked, RateLimitExceeded:
		return "HIGH"
	case LoginFailed, LoginMFAFailed, RegisterFailed, ValidationFailed:
		return "MEDIUM"
	case LoginSuccess, RegisterSuccess, PasswordChanged, MFAEnabled, MFADisabled, TokenRefreshed, LogoutSuccess, AccountUnlocked:
		return "LOW"
	default:
		return "MEDIUM"
	}
}

// GetRecentSecurityEvents retrieves recent high-severity security events
func (al *AuditLogger) GetRecentSecurityEvents(hours int, limit int) ([]AuditEvent, error) {
	if al.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	query := `
		SELECT user_id, email, event_type, ip_address, user_agent, details, severity, timestamp
		FROM audit_logs
		WHERE severity IN ('HIGH', 'CRITICAL') 
		AND timestamp >= DATE_SUB(NOW(), INTERVAL ? HOUR)
		ORDER BY timestamp DESC
		LIMIT ?
	`

	rows, err := al.db.Query(query, hours, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query security events: %v", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var detailsJSON string

		err := rows.Scan(
			&event.UserID,
			&event.Email,
			&event.EventType,
			&event.IPAddress,
			&event.UserAgent,
			&detailsJSON,
			&event.Severity,
			&event.Timestamp,
		)

		if err != nil {
			log.Printf("Error scanning audit event: %v", err)
			continue
		}

		// Parse details JSON
		if detailsJSON != "" {
			json.Unmarshal([]byte(detailsJSON), &event.Details)
		}

		events = append(events, event)
	}

	return events, nil
}

// DetectSuspiciousActivity analyzes patterns to detect potential security threats
func (al *AuditLogger) DetectSuspiciousActivity(ip string, timeWindow time.Duration) bool {
	if al.db == nil {
		return false
	}

	query := `
		SELECT COUNT(*) as failed_attempts
		FROM audit_logs
		WHERE ip_address = ? 
		AND event_type IN ('LOGIN_FAILED', 'VALIDATION_FAILED', 'RATE_LIMIT_EXCEEDED')
		AND timestamp >= DATE_SUB(NOW(), INTERVAL ? MINUTE)
	`

	var failedAttempts int
	err := al.db.QueryRow(query, ip, int(timeWindow.Minutes())).Scan(&failedAttempts)
	if err != nil {
		log.Printf("Error checking suspicious activity: %v", err)
		return false
	}

	// Consider suspicious if more than 10 failed attempts in the time window
	return failedAttempts > 10
}
