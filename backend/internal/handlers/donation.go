package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Donation struct {
	ID               string    `json:"id"`
	DonorID          string    `json:"donorId"`
	DisasterReportID string    `json:"disasterReportId"`
	Amount           float64   `json:"amount"`
	Currency         string    `json:"currency"`
	Description      string    `json:"description"`
	Status           string    `json:"status"`
	TransactionID    string    `json:"transactionId"`
	PaymentMethod    string    `json:"paymentMethod"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

type DonationHandler struct {
	db *sql.DB
}

func NewDonationHandler(db *sql.DB) *DonationHandler {
	return &DonationHandler{db: db}
}

func (h *DonationHandler) CreateDonation(w http.ResponseWriter, r *http.Request) {
	var donation struct {
		DisasterReportID string  `json:"disasterReportId"`
		Amount           float64 `json:"amount"`
		Currency         string  `json:"currency"`
		Description      string  `json:"description"`
		PaymentMethod    string  `json:"paymentMethod"`
	}

	if err := json.NewDecoder(r.Body).Decode(&donation); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate amount
	if donation.Amount <= 0 {
		http.Error(w, "Invalid donation amount", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Verify disaster report exists and is verified
	var reportStatus string
	err = tx.QueryRow(
		"SELECT status FROM disaster_reports WHERE id = UUID_TO_BIN(?) FOR UPDATE",
		donation.DisasterReportID,
	).Scan(&reportStatus)

	if err == sql.ErrNoRows {
		http.Error(w, "Disaster report not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Error verifying disaster report", http.StatusInternalServerError)
		return
	}

	if reportStatus != "verified" {
		http.Error(w, "Cannot donate to unverified disaster report", http.StatusBadRequest)
		return
	}

	// Get user ID from context
	userID := r.Context().Value("user_id").(string)

	// Generate transaction ID
	transactionID := generateTransactionID()

	// Insert donation
	var donationID string
	err = tx.QueryRow(
		`INSERT INTO donations (
			id, donor_id, disaster_report_id, amount, currency, 
			description, status, transaction_id, payment_method
		) VALUES (
			UUID_TO_BIN(UUID()), UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?, 
			?, 'pending', ?, ?
		) RETURNING BIN_TO_UUID(id)`,
		userID, donation.DisasterReportID, donation.Amount, donation.Currency,
		donation.Description, transactionID, donation.PaymentMethod,
	).Scan(&donationID)

	if err != nil {
		http.Error(w, "Error creating donation", http.StatusInternalServerError)
		return
	}

	// Insert audit log
	_, err = tx.Exec(
		`INSERT INTO audit_logs (
			id, user_id, action, entity_type, entity_id, 
			ip_address, user_agent, details
		) VALUES (
			UUID_TO_BIN(UUID()), UUID_TO_BIN(?), 'create_donation', 'donation', 
			UUID_TO_BIN(?), ?, ?, ?
		)`,
		userID, donationID, r.RemoteAddr, r.UserAgent(),
		json.RawMessage(`{"amount":"`+fmt.Sprintf("%.2f", donation.Amount)+`","currency":"`+donation.Currency+`"}`),
	)

	if err != nil {
		http.Error(w, "Error logging donation", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error finalizing donation", http.StatusInternalServerError)
		return
	}

	// Return donation details
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":            donationID,
		"transactionId": transactionID,
		"status":        "pending",
		"message":       "Donation created successfully",
	})
}

func (h *DonationHandler) GetDonation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	donationID := vars["id"]
	userID := r.Context().Value("user_id").(string)

	var donation Donation
	err := h.db.QueryRow(
		`SELECT BIN_TO_UUID(id), BIN_TO_UUID(donor_id), BIN_TO_UUID(disaster_report_id),
		amount, currency, description, status, transaction_id, payment_method,
		created_at, updated_at
		FROM donations 
		WHERE id = UUID_TO_BIN(?) AND (donor_id = UUID_TO_BIN(?) OR 
		disaster_report_id IN (
			SELECT id FROM disaster_reports WHERE reporter_id = UUID_TO_BIN(?)
		))`,
		donationID, userID, userID,
	).Scan(
		&donation.ID, &donation.DonorID, &donation.DisasterReportID,
		&donation.Amount, &donation.Currency, &donation.Description,
		&donation.Status, &donation.TransactionID, &donation.PaymentMethod,
		&donation.CreatedAt, &donation.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Donation not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Error fetching donation", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(donation)
}

func (h *DonationHandler) ListDonations(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(string)

	// Parse query parameters
	limit := 10
	offset := 0
	status := r.URL.Query().Get("status")
	reportID := r.URL.Query().Get("reportId")

	query := `
		SELECT BIN_TO_UUID(d.id), BIN_TO_UUID(d.donor_id), BIN_TO_UUID(d.disaster_report_id),
		d.amount, d.currency, d.description, d.status, d.transaction_id, d.payment_method,
		d.created_at, d.updated_at
		FROM donations d
		WHERE (d.donor_id = UUID_TO_BIN(?) OR 
		d.disaster_report_id IN (
			SELECT id FROM disaster_reports WHERE reporter_id = UUID_TO_BIN(?)
		))`

	args := []interface{}{userID, userID}

	if status != "" {
		query += " AND d.status = ?"
		args = append(args, status)
	}
	if reportID != "" {
		query += " AND d.disaster_report_id = UUID_TO_BIN(?)"
		args = append(args, reportID)
	}

	query += " ORDER BY d.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Error fetching donations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var donations []Donation
	for rows.Next() {
		var d Donation
		if err := rows.Scan(
			&d.ID, &d.DonorID, &d.DisasterReportID,
			&d.Amount, &d.Currency, &d.Description,
			&d.Status, &d.TransactionID, &d.PaymentMethod,
			&d.CreatedAt, &d.UpdatedAt,
		); err != nil {
			http.Error(w, "Error processing donations", http.StatusInternalServerError)
			return
		}
		donations = append(donations, d)
	}

	json.NewEncoder(w).Encode(donations)
}

func (h *DonationHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	donationID := vars["id"]
	userID := r.Context().Value("user_id").(string)

	var update struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update donation status
	result, err := tx.Exec(
		`UPDATE donations 
		SET status = ?, updated_at = NOW()
		WHERE id = UUID_TO_BIN(?) AND donor_id = UUID_TO_BIN(?)`,
		update.Status, donationID, userID,
	)

	if err != nil {
		http.Error(w, "Error updating donation status", http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking update result", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "Donation not found or unauthorized", http.StatusNotFound)
		return
	}

	// Log the status update
	_, err = tx.Exec(
		`INSERT INTO audit_logs (
			id, user_id, action, entity_type, entity_id, 
			ip_address, user_agent, details
		) VALUES (
			UUID_TO_BIN(UUID()), UUID_TO_BIN(?), 'update_donation_status', 
			'donation', UUID_TO_BIN(?), ?, ?, ?
		)`,
		userID, donationID, r.RemoteAddr, r.UserAgent(),
		json.RawMessage(`{"status":"`+update.Status+`"}`),
	)

	if err != nil {
		http.Error(w, "Error logging status update", http.StatusInternalServerError)
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error finalizing status update", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Donation status updated successfully",
	})
}

func generateTransactionID() string {
	timestamp := time.Now().Format("20060102150405")
	random := make([]byte, 4)
	rand.Read(random)
	return fmt.Sprintf("TRX-%s-%x", timestamp, random)
}
