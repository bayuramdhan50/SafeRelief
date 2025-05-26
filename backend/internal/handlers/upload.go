package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type Upload struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"originalName"`
	Size         int64     `json:"size"`
	MimeType     string    `json:"mimeType"`
	Path         string    `json:"path"`
	CreatedAt    time.Time `json:"createdAt"`
}

type UploadHandler struct {
	db        *sql.DB
	uploadDir string
}

func NewUploadHandler(db *sql.DB) *UploadHandler {
	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, 0755)
	return &UploadHandler{
		db:        db,
		uploadDir: uploadDir,
	}
}

func (h *UploadHandler) UploadFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(string)

	// Parse multipart form
	err := r.ParseMultipartForm(25 << 20) // 25MB max
	if err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["files"]
	var uploads []Upload

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Failed to open file", http.StatusInternalServerError)
			return
		}
		defer file.Close()

		// Validate file type
		if !h.isAllowedFileType(fileHeader.Filename) {
			http.Error(w, fmt.Sprintf("File type not allowed: %s", fileHeader.Filename), http.StatusBadRequest)
			return
		}

		// Validate file size
		if fileHeader.Size > maxFileSize {
			http.Error(w, "File too large", http.StatusBadRequest)
			return
		}

		// Generate unique filename
		filename := h.generateUniqueFilename(fileHeader.Filename)
		filePath := filepath.Join(h.uploadDir, filename)

		// Save file to disk
		dst, err := os.Create(filePath)
		if err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, "Failed to save file", http.StatusInternalServerError)
			return
		}

		// Save to database
		upload := Upload{
			ID:           h.generateID(),
			UserID:       userID,
			Filename:     filename,
			OriginalName: fileHeader.Filename,
			Size:         fileHeader.Size,
			MimeType:     fileHeader.Header.Get("Content-Type"),
			Path:         filePath,
			CreatedAt:    time.Now(),
		}

		_, err = h.db.Exec(`
			INSERT INTO uploads (id, user_id, filename, original_name, size, mime_type, path, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, upload.ID, upload.UserID, upload.Filename, upload.OriginalName,
			upload.Size, upload.MimeType, upload.Path, upload.CreatedAt)

		if err != nil {
			// Clean up file if database insert fails
			os.Remove(filePath)
			http.Error(w, "Failed to save upload record", http.StatusInternalServerError)
			return
		}

		uploads = append(uploads, upload)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Files uploaded successfully",
		"uploads": uploads,
	})
}

func (h *UploadHandler) GetFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fileID := vars["id"]

	var upload Upload
	err := h.db.QueryRow(`
		SELECT id, user_id, filename, original_name, size, mime_type, path, created_at
		FROM uploads WHERE id = ?
	`, fileID).Scan(&upload.ID, &upload.UserID, &upload.Filename, &upload.OriginalName,
		&upload.Size, &upload.MimeType, &upload.Path, &upload.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Check if file exists on disk
	if _, err := os.Stat(upload.Path); os.IsNotExist(err) {
		http.Error(w, "File not found on disk", http.StatusNotFound)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", upload.MimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", upload.OriginalName))

	// Serve file
	http.ServeFile(w, r, upload.Path)
}

func (h *UploadHandler) isAllowedFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".doc", ".docx"}

	for _, allowed := range allowedExts {
		if ext == allowed {
			return true
		}
	}
	return false
}

func (h *UploadHandler) generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d_%s%s", timestamp, h.generateID()[:8], ext)
}

func (h *UploadHandler) generateID() string {
	// Simple ID generation - in production, use a proper UUID library
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
