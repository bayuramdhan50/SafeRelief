package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/didip/tollbooth"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/unrolled/secure"

	"saferelief/internal/middleware"
)

func main() {
	// Load .env file from backend root directory
	if err := godotenv.Load("../../.env"); err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
		log.Println("Continuing with environment variables...")
	}

	// Database connection
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "saferelief_db")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()
	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize router
	router := mux.NewRouter()

	// CORS middleware - MUST be first to handle preflight requests
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Allow localhost:3000 for development
			if origin == "http://localhost:3000" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, Accept, Origin, User-Agent")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
	// Apply other security middleware in order
	router.Use(middleware.SecurityHeaders)
	router.Use(middleware.InputSanitizationMiddleware)
	router.Use(middleware.RateLimitMiddleware(100, time.Hour)) // 100 requests per hour
	// Note: CSRF will be applied selectively to protected routes only

	// Enhanced security middleware
	secureMiddleware := secure.New(secure.Options{
		AllowedHosts:          []string{"localhost:3000", "localhost:8080"},
		SSLRedirect:           false, // Set to true in production
		SSLHost:               "localhost:8080",
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		BrowserXssFilter:      true,
		ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com",
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		PermissionsPolicy:     "camera=(), microphone=(), geolocation=()",
	})

	// Apply additional security headers
	router.Use(func(next http.Handler) http.Handler {
		return secureMiddleware.Handler(next)
	})

	// Legacy rate limiting (keeping for compatibility)
	lmt := tollbooth.NewLimiter(1, nil)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			httpError := tollbooth.LimitByRequest(lmt, w, r)
			if httpError != nil {
				http.Error(w, httpError.Message, httpError.StatusCode)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// Create API router with /api prefix
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Setup all routes from routes.go (includes auth and protected routes)
	setupRoutes(apiRouter, db)

	// Health check endpoint (public, not under /api)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
	}).Methods("GET", "OPTIONS")

	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s with enhanced security...", port)
	log.Printf("Health check available at: http://localhost:%s/health", port)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Fatal(server.ListenAndServe())
}

// getEnv gets environment variable with fallback
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
