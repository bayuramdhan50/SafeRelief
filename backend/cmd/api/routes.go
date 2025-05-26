package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"saferelief/internal/auth"
	"saferelief/internal/handlers"
	"saferelief/internal/middleware"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func initDB() (*sql.DB, error) {
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	return db, db.Ping()
}

func setupRoutes() *mux.Router {
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	refreshSecret := []byte(os.Getenv("REFRESH_TOKEN_SECRET"))
	csrfSecret := []byte(os.Getenv("CSRF_SECRET"))

	// Initialize handlers
	authHandler := auth.NewAuthHandler(jwtSecret, refreshSecret, db)
	reportHandler := handlers.NewReportHandler(db)
	donationHandler := handlers.NewDonationHandler(db)
	userHandler := handlers.NewUserHandler(db)
	uploadHandler := handlers.NewUploadHandler(db)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)
	csrfMiddleware := middleware.NewCSRFMiddleware(csrfSecret)

	// Create main router
	router := mux.NewRouter()

	// Router configuration
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Apply global middleware
	apiRouter.Use(middleware.SecurityHeaders)
	apiRouter.Use(middleware.SanitizeInput)
	apiRouter.Use(csrfMiddleware.ValidateCSRF)
	// Auth routes
	authRouter := apiRouter.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST")
	authRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST")
	authRouter.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST")

	// Protected routes
	protectedRouter := apiRouter.PathPrefix("").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)

	// User routes
	protectedRouter.HandleFunc("/users/me", userHandler.GetProfile).Methods("GET")
	protectedRouter.HandleFunc("/users/me", userHandler.UpdateProfile).Methods("PUT")
	protectedRouter.HandleFunc("/users/me/mfa", userHandler.EnableMFA).Methods("POST")
	protectedRouter.HandleFunc("/users/me/mfa", userHandler.DisableMFA).Methods("DELETE")

	// Disaster report routes
	protectedRouter.HandleFunc("/reports", reportHandler.CreateReport).Methods("POST")
	protectedRouter.HandleFunc("/reports", reportHandler.ListReports).Methods("GET")
	protectedRouter.HandleFunc("/reports/{id}", reportHandler.GetReport).Methods("GET")
	protectedRouter.HandleFunc("/reports/{id}", reportHandler.UpdateReport).Methods("PUT")
	protectedRouter.HandleFunc("/reports/{id}/verify", reportHandler.VerifyReport).Methods("POST")

	// Donation routes
	protectedRouter.HandleFunc("/donations", donationHandler.CreateDonation).Methods("POST")
	protectedRouter.HandleFunc("/donations", donationHandler.ListDonations).Methods("GET")
	protectedRouter.HandleFunc("/donations/{id}", donationHandler.GetDonation).Methods("GET")
	protectedRouter.HandleFunc("/donations/{id}/status", donationHandler.UpdateStatus).Methods("PUT")

	// File upload routes with specific security measures
	protectedRouter.HandleFunc("/uploads", uploadHandler.UploadFiles).Methods("POST")
	protectedRouter.HandleFunc("/uploads/{id}", uploadHandler.GetFile).Methods("GET")

	return router
}
