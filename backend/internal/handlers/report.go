package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	maxFileSize  = 5 * 1024 * 1024  // 5MB
	maxTotalSize = 25 * 1024 * 1024 // 25MB
	allowedTypes = ".jpg,.jpeg,.png"
	uploadDir    = "./uploads"
)

type DisasterReport struct {
	ID          string    `json:"id"`
	ReporterID  string    `json:"reporterId"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Severity    string    `json:"severity"`
	Status      string    `json:"status"`
	VerifiedBy  *string   `json:"verifiedBy"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Files       []File    `json:"files,omitempty"`
}

type File struct {
	ID        string    `json:"id"`
	Filename  string    `json:"filename"`
	FileHash  string    `json:"fileHash"`
	FileSize  int64     `json:"fileSize"`
	MimeType  string    `json:"mimeType"`
	CreatedAt time.Time `json:"createdAt"`
}

type ReportHandler struct {
	db *sql.DB
}

func NewReportHandler(db *sql.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

func (h *ReportHandler) CreateReport(w http.ResponseWriter, r *http.Request) {
	// Limit the entire request body size
	r.Body = http.MaxBytesReader(w, r.Body, maxTotalSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxTotalSize); err != nil {
		http.Error(w, "Request too large", http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	// Get user ID from context
	userID := r.Context().Value("user_id").(string)

	// Start transaction
	tx, err := h.db.Begin()
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert report
	var reportID string
	err = tx.QueryRow(
		`INSERT INTO disaster_reports (id, reporter_id, title, description, latitude, longitude, severity, status)
		VALUES (UUID_TO_BIN(UUID()), UUID_TO_BIN(?), ?, ?, ?, ?, ?, 'pending')
		RETURNING BIN_TO_UUID(id)`,
		userID,
		r.FormValue("title"),
		r.FormValue("description"),
		r.FormValue("latitude"),
		r.FormValue("longitude"),
		r.FormValue("severity"),
	).Scan(&reportID)

	if err != nil {
		http.Error(w, "Error creating report", http.StatusInternalServerError)
		return
	}

	// Handle file uploads
	files := r.MultipartForm.File["files"]
	for _, fileHeader := range files {
		if err := h.validateAndSaveFile(tx, reportID, userID, fileHeader); err != nil {
			http.Error(w, "Error processing file upload", http.StatusBadRequest)
			return
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error saving report", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id":      reportID,
		"message": "Report created successfully",
	})
}

func (h *ReportHandler) validateAndSaveFile(tx *sql.Tx, reportID, userID string, fileHeader *multipart.FileHeader) error {
	// Check file size
	if fileHeader.Size > maxFileSize {
		return fmt.Errorf("file too large")
	}

	// Check file type
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !strings.Contains(allowedTypes, ext) {
		return fmt.Errorf("invalid file type")
	}

	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	// Calculate file hash
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	fileHash := hex.EncodeToString(hash.Sum(nil))

	// Reset file pointer
	file.Seek(0, 0)

	// Create unique filename
	filename := fmt.Sprintf("%s-%s%s", reportID, fileHash[:8], ext)
	filepath := filepath.Join(uploadDir, filename)

	// Create destination file
	dst, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy file contents
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filepath)
		return err
	}

	// Insert file record
	_, err = tx.Exec(
		`INSERT INTO file_uploads (id, user_id, disaster_report_id, filename, original_filename, file_size, mime_type, file_hash, storage_path)
		VALUES (UUID_TO_BIN(UUID()), UUID_TO_BIN(?), UUID_TO_BIN(?), ?, ?, ?, ?, ?, ?)`,
		userID, reportID, filename, fileHeader.Filename, fileHeader.Size, fileHeader.Header.Get("Content-Type"), fileHash, filepath,
	)

	return err
}

func (h *ReportHandler) GetReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]

	var report DisasterReport
	err := h.db.QueryRow(
		`SELECT BIN_TO_UUID(id), BIN_TO_UUID(reporter_id), title, description, 
		latitude, longitude, severity, status, BIN_TO_UUID(verified_by), created_at, updated_at
		FROM disaster_reports WHERE id = UUID_TO_BIN(?)`,
		reportID,
	).Scan(
		&report.ID, &report.ReporterID, &report.Title, &report.Description,
		&report.Latitude, &report.Longitude, &report.Severity, &report.Status,
		&report.VerifiedBy, &report.CreatedAt, &report.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Error fetching report", http.StatusInternalServerError)
		return
	}

	// Get associated files
	rows, err := h.db.Query(
		`SELECT BIN_TO_UUID(id), filename, file_hash, file_size, mime_type, created_at
		FROM file_uploads WHERE disaster_report_id = UUID_TO_BIN(?)`,
		reportID,
	)
	if err != nil {
		http.Error(w, "Error fetching files", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var file File
		if err := rows.Scan(&file.ID, &file.Filename, &file.FileHash, &file.FileSize, &file.MimeType, &file.CreatedAt); err != nil {
			http.Error(w, "Error processing files", http.StatusInternalServerError)
			return
		}
		report.Files = append(report.Files, file)
	}

	json.NewEncoder(w).Encode(report)
}

func (h *ReportHandler) ListReports(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering and pagination
	limit := 10
	offset := 0
	status := r.URL.Query().Get("status")
	severity := r.URL.Query().Get("severity")

	query := `SELECT BIN_TO_UUID(id), BIN_TO_UUID(reporter_id), title, description, 
		latitude, longitude, severity, status, BIN_TO_UUID(verified_by), created_at, updated_at
		FROM disaster_reports WHERE 1=1`
	args := []interface{}{}

	if status != "" {
		query += " AND status = ?"
		args = append(args, status)
	}
	if severity != "" {
		query += " AND severity = ?"
		args = append(args, severity)
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Error fetching reports", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []DisasterReport
	for rows.Next() {
		var report DisasterReport
		if err := rows.Scan(
			&report.ID, &report.ReporterID, &report.Title, &report.Description,
			&report.Latitude, &report.Longitude, &report.Severity, &report.Status,
			&report.VerifiedBy, &report.CreatedAt, &report.UpdatedAt,
		); err != nil {
			http.Error(w, "Error processing reports", http.StatusInternalServerError)
			return
		}
		reports = append(reports, report)
	}

	json.NewEncoder(w).Encode(reports)
}

func (h *ReportHandler) VerifyReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]
	userID := r.Context().Value("user_id").(string)

	// Update report status
	result, err := h.db.Exec(
		`UPDATE disaster_reports 
		SET status = 'verified', verified_by = UUID_TO_BIN(?), updated_at = NOW()
		WHERE id = UUID_TO_BIN(?) AND status = 'pending'`,
		userID, reportID,
	)
	if err != nil {
		http.Error(w, "Error verifying report", http.StatusInternalServerError)
		return
	}

	rows, err := result.RowsAffected()
	if err != nil {
		http.Error(w, "Error checking update result", http.StatusInternalServerError)
		return
	}
	if rows == 0 {
		http.Error(w, "Report not found or already verified", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Report verified successfully",
	})
}
