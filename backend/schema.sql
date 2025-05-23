-- Create database with secure defaults
CREATE DATABASE IF NOT EXISTS saferelief_db
CHARACTER SET utf8mb4
COLLATE utf8mb4_unicode_ci;

USE saferelief_db;

-- Users table with security features
CREATE TABLE IF NOT EXISTS users (
    id BINARY(16) PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash CHAR(60) NOT NULL,
    mfa_secret VARCHAR(32),
    mfa_enabled BOOLEAN DEFAULT FALSE,
    failed_attempts INT DEFAULT 0,
    locked_until DATETIME,
    last_password_change DATETIME NOT NULL,
    require_password_change BOOLEAN DEFAULT FALSE,
    status ENUM('active', 'inactive', 'banned') DEFAULT 'inactive',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_username (username)
) ENGINE=InnoDB;

-- Sessions table for secure session management
CREATE TABLE IF NOT EXISTS sessions (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16) NOT NULL,
    token_hash CHAR(64) NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_token_hash (token_hash),
    INDEX idx_expires_at (expires_at)
) ENGINE=InnoDB;

-- Disaster reports with location data
CREATE TABLE IF NOT EXISTS disaster_reports (
    id BINARY(16) PRIMARY KEY,
    reporter_id BINARY(16) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    latitude DECIMAL(10,8) NOT NULL,
    longitude DECIMAL(11,8) NOT NULL,
    location POINT NOT NULL SRID 4326,
    severity ENUM('low', 'medium', 'high', 'critical') NOT NULL,
    status ENUM('pending', 'verified', 'resolved') DEFAULT 'pending',
    verified_by BINARY(16),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (reporter_id) REFERENCES users(id),
    FOREIGN KEY (verified_by) REFERENCES users(id),
    INDEX idx_status (status),
    INDEX idx_coords (latitude, longitude),
    SPATIAL INDEX idx_location (location)
) ENGINE=InnoDB;

-- Create trigger to set POINT data from latitude and longitude
DELIMITER //
CREATE TRIGGER disaster_reports_before_insert 
BEFORE INSERT ON disaster_reports
FOR EACH ROW
BEGIN
    SET NEW.location = ST_SRID(POINT(NEW.longitude, NEW.latitude), 4326);
END//

CREATE TRIGGER disaster_reports_before_update
BEFORE UPDATE ON disaster_reports
FOR EACH ROW
BEGIN
    IF NEW.latitude != OLD.latitude OR NEW.longitude != OLD.longitude THEN
        SET NEW.location = ST_SRID(POINT(NEW.longitude, NEW.latitude), 4326);
    END IF;
END//
DELIMITER ;

-- Donations with transaction tracking
CREATE TABLE IF NOT EXISTS donations (
    id BINARY(16) PRIMARY KEY,
    donor_id BINARY(16) NOT NULL,
    disaster_report_id BINARY(16) NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    currency CHAR(3) NOT NULL DEFAULT 'IDR',
    description TEXT,
    status ENUM('pending', 'completed', 'failed', 'refunded') DEFAULT 'pending',
    transaction_id VARCHAR(100),
    payment_method VARCHAR(50),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (donor_id) REFERENCES users(id),
    FOREIGN KEY (disaster_report_id) REFERENCES disaster_reports(id),
    INDEX idx_status (status),
    INDEX idx_transaction (transaction_id)
) ENGINE=InnoDB;

-- Audit logs for security tracking
CREATE TABLE IF NOT EXISTS audit_logs (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BINARY(16),
    ip_address VARCHAR(45) NOT NULL,
    user_agent VARCHAR(255),
    details JSON,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_action (action),
    INDEX idx_entity (entity_type, entity_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB;

-- Rate limiting table
CREATE TABLE IF NOT EXISTS rate_limits (
    id BINARY(16) PRIMARY KEY,
    ip_address VARCHAR(45) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    request_count INT NOT NULL DEFAULT 1,
    window_start DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_ip_endpoint (ip_address, endpoint),
    INDEX idx_window (window_start)
) ENGINE=InnoDB;

-- File uploads tracking
CREATE TABLE IF NOT EXISTS file_uploads (
    id BINARY(16) PRIMARY KEY,
    user_id BINARY(16) NOT NULL,
    disaster_report_id BINARY(16) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_size INT NOT NULL,
    mime_type VARCHAR(127) NOT NULL,
    file_hash CHAR(64) NOT NULL,
    storage_path VARCHAR(512) NOT NULL,
    status ENUM('pending', 'verified', 'rejected') DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (disaster_report_id) REFERENCES disaster_reports(id),
    INDEX idx_file_hash (file_hash),
    INDEX idx_status (status)
) ENGINE=InnoDB;

-- Create secure user for application
CREATE USER IF NOT EXISTS 'saferelief_user'@'localhost' IDENTIFIED BY 'your-strong-password-here';
GRANT SELECT, INSERT, UPDATE, DELETE ON saferelief_db.* TO 'saferelief_user'@'localhost';
FLUSH PRIVILEGES;
