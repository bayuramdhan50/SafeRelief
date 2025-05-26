package main

import (
	"database/sql"
	"fmt"
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

func setupRoutes(router *mux.Router, db *sql.DB) {
	jwtSecret := []byte(os.Getenv("JWT_SECRET"))
	refreshSecret := []byte(os.Getenv("REFRESH_TOKEN_SECRET"))
	// Initialize handlers
	authHandler := auth.NewAuthHandler(jwtSecret, refreshSecret, db)
	reportHandler := handlers.NewReportHandler(db)
	donationHandler := handlers.NewDonationHandler(db)
	userHandler := handlers.NewUserHandler(db)
	uploadHandler := handlers.NewUploadHandler(db)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtSecret)

	// Router is already /api prefixed, so we don't need to add it again

	// Auth routes (public, no authentication required)
	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/register", authHandler.Register).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/login", authHandler.Login).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	authRouter.HandleFunc("/refresh", authHandler.RefreshToken).Methods("POST", "OPTIONS")
	// Protected routes
	protectedRouter := router.PathPrefix("").Subrouter()
	protectedRouter.Use(authMiddleware.Authenticate)

	// User routes
	protectedRouter.HandleFunc("/users/me", userHandler.GetProfile).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/users/me", userHandler.UpdateProfile).Methods("PUT", "OPTIONS")
	protectedRouter.HandleFunc("/users/me/mfa", userHandler.EnableMFA).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/users/me/mfa", userHandler.DisableMFA).Methods("DELETE", "OPTIONS")

	// Disaster report routes
	protectedRouter.HandleFunc("/reports", reportHandler.CreateReport).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/reports", reportHandler.ListReports).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/reports/{id}", reportHandler.GetReport).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/reports/{id}", reportHandler.UpdateReport).Methods("PUT", "OPTIONS")
	protectedRouter.HandleFunc("/reports/{id}/verify", reportHandler.VerifyReport).Methods("POST", "OPTIONS")

	// Donation routes
	protectedRouter.HandleFunc("/donations", donationHandler.CreateDonation).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/donations", donationHandler.ListDonations).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/donations/{id}", donationHandler.GetDonation).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/donations/{id}/status", donationHandler.UpdateStatus).Methods("PUT", "OPTIONS")
	// File upload routes with specific security measures
	protectedRouter.HandleFunc("/uploads", uploadHandler.UploadFiles).Methods("POST", "OPTIONS")
	protectedRouter.HandleFunc("/uploads/{id}", uploadHandler.GetFile).Methods("GET", "OPTIONS")
}
