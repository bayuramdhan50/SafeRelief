# 🚨 SafeRelief - Platform Donasi Bantuan Bencana

<div align="center">

**Platform donasi bantuan bencana yang aman, transparan, dan terpercaya untuk Indonesia**

[![Next.js](https://img.shields.io/badge/Next.js-15.1.8-black)](https://nextjs.org/)
[![Go](https://img.shields.io/badge/Go-1.23+-blue)](https://golang.org/)
[![MySQL](https://img.shields.io/badge/MySQL-8.0+-orange)](https://mysql.com/)
[![Security](https://img.shields.io/badge/Security-Enterprise-green)](https://github.com)

</div>

## 🎯 Tentang SafeRelief

SafeRelief adalah platform inovatif yang menghubungkan kebaikan hati masyarakat dengan korban bencana alam di seluruh Indonesia. Dengan fokus pada keamanan tingkat enterprise dan transparansi penuh, kami memastikan setiap rupiah donasi sampai ke tangan yang membutuhkan.

### ✨ Fitur Utama

- 🔐 **Keamanan Tinggi** - Multi-layer security dengan JWT authentication
- 🗺️ **Peta Interaktif** - Visualisasi lokasi bencana dengan React Leaflet 
- 📱 **Responsive Design** - Dioptimalkan untuk semua perangkat
- 🚀 **Performa Tinggi** - Dibangun dengan Next.js 15 dan Go
- 📊 **Dashboard Transparan** - Tracking donasi real-time
- 🎨 **UI/UX Modern** - Desain modern dengan Tailwind CSS
- � **Form Validasi** - Validasi form dengan React Hook Form + Zod
- 🔔 **Notifikasi Real-time** - Toast notifications dengan React Toastify

### 🏆 Mendukung SDGs

- **SDG 11**: Kota dan Komunitas Berkelanjutan
- **SDG 13**: Aksi Iklim dan Mitigasi Bencana
- **SDG 1**: Tanpa Kemiskinan
- **SDG 17**: Kemitraan untuk Mencapai Tujuan

## 🏗️ Arsitektur Sistem

SafeRelief dibangun dengan arsitektur modern yang mengutamakan keamanan, performa, dan skalabilitas.

### 🎨 Frontend (Next.js 15)
```
📦 Tech Stack
├── ⚛️  Next.js 15.1.8 (App Router)
├── 🎨 Tailwind CSS 3.4.1
├── 🗺️  React Leaflet 5.0.0 + Leaflet 1.9.4
├── 📝 React Hook Form 7.56.4 + Zod 3.25.28
├── 🔔 React Toastify 11.0.5
├── � QR Code React 4.2.0
├── 🔒 JWT Authentication
└── 🛡️  Security Middleware
```

**Fitur Keamanan Frontend:**
- ✅ Server-side rendering (SSR) untuk SEO dan keamanan
- ✅ Input validation & sanitization dengan Zod
- ✅ Security headers dengan middleware
- ✅ Form validation dengan React Hook Form
- ✅ Type safety dengan TypeScript 5

### ⚡ Backend (Golang)
```
📦 Tech Stack
├── 🚀 Go 1.23+ (Gorilla Mux)
├── 🗄️  MySQL 8.0+ (Native Driver)
├── 🔐 JWT Authentication (golang-jwt/jwt/v5)
├── � Password Hashing (golang.org/x/crypto)
├── � Rate Limiting (tollbooth)
├── �️  Security Headers (unrolled/secure)
├── � Multi-Factor Auth (pquerna/otp)
└── ⚙️  Environment Config (godotenv)
```

**Fitur Keamanan Backend:**
- ✅ Clean Architecture dengan modular structure
- ✅ JWT authentication dengan RS256 signature
- ✅ Bcrypt password hashing
- ✅ Rate limiting middleware
- ✅ Input validation dan SQL injection prevention
- ✅ File upload validation
- ✅ Audit logging untuk security events
- ✅ MFA support dengan TOTP

### 🗄️ Database (MySQL 8.0)
```
📊 Database Features
├── 🔒 Secure connections dengan proper authentication
├── 🔐 Encrypted sensitive data storage
├── 📝 Audit logging untuk semua transaksi
├── � Optimized queries dengan proper indexing
├── 👥 Role-based access control
├── � Spatial indexing untuk location queries
└── � UUID primary keys untuk security
```

## 🛡️ Fitur Keamanan

SafeRelief menerapkan security-first approach dengan multiple layers of protection.

### 🔐 Authentication & Authorization
```
🔑 Security Features
├── 🔑 JWT Authentication dengan RSA-256
├── 🔒 Bcrypt password hashing
├── 🚫 Account lockout setelah failed attempts
├── 👥 Role-based Access Control (RBAC)
├── 🕐 Session timeout management
├── 📱 Multi-Factor Authentication support
├── � Audit logging untuk authentication events
└── 🔍 Rate limiting untuk API endpoints
```

### 🛡️ Protection Methods
| Security Aspect | Implementation | Status |
|-----------------|----------------|--------|
| **Authentication** | JWT + MFA | ✅ Implemented |
| **Input Validation** | Zod + Backend validation | ✅ Implemented |
| **SQL Injection** | Prepared statements | ✅ Implemented |
| **XSS Protection** | Input sanitization | ✅ Implemented |
| **CSRF Protection** | Security headers | ✅ Implemented |
| **Rate Limiting** | Tollbooth middleware | ✅ Implemented |
| **File Upload Security** | Type & size validation | ✅ Implemented |
| **Session Management** | Secure JWT handling | ✅ Implemented |

### 🔒 Security Headers (Middleware)
```http
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

## 📊 Database Schema

### 👥 Users Table
```sql
CREATE TABLE users (
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
);
```

### � Disaster Reports Table
```sql
CREATE TABLE disaster_reports (
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
);
```

### � Donations Table
```sql
CREATE TABLE donations (
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
);
```

## 🚀 API Endpoints

### 🔐 Authentication
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout
- `POST /api/auth/mfa/setup` - Setup MFA
- `POST /api/auth/mfa/verify` - Verify MFA token

### 💰 Donations
- `POST /api/donations` - Create donation
- `GET /api/donations` - List donations
- `GET /api/donations/:id` - Get donation details
- `PATCH /api/donations/:id/status` - Update donation status

### 🚨 Disaster Reports
- `POST /api/reports` - Create disaster report
- `GET /api/reports` - List disaster reports
- `GET /api/reports/:id` - Get report details
- `PATCH /api/reports/:id/verify` - Verify report (admin only)
- `POST /api/reports/:id/upload` - Upload evidence files

## 🚀 Quick Start

### 📋 Prerequisites
```bash
# Required Software
✅ Node.js 18+ dan npm
✅ Go 1.23+
✅ MySQL 8.0+
✅ Git
```

### 🛠️ Installation

#### 1️⃣ Clone Repository
```bash
git clone https://github.com/username/saferelief.git
cd saferelief
```

#### 2️⃣ Database Setup
```bash
# Start MySQL service
# Windows: net start mysql
# Linux/Mac: sudo systemctl start mysql

# Create database dan user
mysql -u root -p < backend/schema.sql
```

#### 3️⃣ Backend Setup
```bash
cd backend

# Install dependencies
go mod download

# Setup environment variables
copy .env.example .env
# Edit .env dengan konfigurasi database Anda

# Run the API server
go run cmd/api/main.go
```

#### 4️⃣ Frontend Setup
```bash
cd frontend

# Install dependencies
npm install

# Setup environment variables
copy .env.example .env.local
# Edit .env.local dengan konfigurasi API endpoint

# Run development server
npm run dev
```

### 🌐 Access Application
```
Frontend: http://localhost:3000
Backend API: http://localhost:8080
```

### 🌐 Environment Variables

#### Frontend (.env.local)
```env
# API Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080

# Map Configuration (optional)
NEXT_PUBLIC_MAPBOX_TOKEN=your_mapbox_token

# App Configuration
NEXT_PUBLIC_APP_NAME=SafeRelief
NEXT_PUBLIC_APP_URL=http://localhost:3000
```

#### Backend (.env)
```env
# Server Configuration
PORT=8080
HOST=localhost

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=saferelief_user
DB_PASSWORD=your-strong-password-here
DB_NAME=saferelief_db

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_EXPIRY=24h

# Security Configuration
BCRYPT_COST=12
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=3600

# File Upload Configuration
MAX_FILE_SIZE=10485760
UPLOAD_DIR=./uploads
```

## 🏗️ Project Structure

```
SafeRelief/
├── frontend/                 # Next.js Frontend
│   ├── app/                 # App Router pages
│   │   ├── components/      # Reusable components
│   │   ├── contexts/        # React contexts
│   │   ├── hooks/          # Custom hooks
│   │   ├── dashboard/      # Dashboard pages
│   │   ├── disasters/      # Disaster management
│   │   ├── donate/         # Donation pages
│   │   ├── login/          # Authentication
│   │   └── register/       # User registration
│   ├── public/             # Static assets
│   └── middleware.ts       # Next.js middleware
├── backend/                # Go Backend
│   ├── cmd/               # Application entry points
│   │   └── api/           # API server
│   ├── internal/          # Internal packages
│   │   ├── auth/          # Authentication logic
│   │   ├── handlers/      # HTTP handlers
│   │   └── middleware/    # HTTP middleware
│   ├── uploads/           # File uploads
│   └── schema.sql         # Database schema
└── README.md              # This file
```

## 🧪 Testing & Development

### 🔧 Available Scripts

#### Frontend
```bash
npm run dev          # Start development server
npm run build        # Build for production
npm run start        # Start production server
npm run lint         # Run ESLint
```

#### Backend
```bash
go run cmd/api/main.go       # Start API server
go test ./...                # Run tests
go build -o bin/api cmd/api/main.go  # Build binary
```

### 🔍 Security Testing

Regular security testing includes:
- ✅ Input validation testing
- ✅ Authentication & authorization testing
- ✅ SQL injection prevention testing
- ✅ XSS protection testing
- ✅ File upload security testing
- ✅ Rate limiting verification
- ✅ Session management testing

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Team

- **Frontend Developer** - Next.js, React, Tailwind CSS
- **Backend Developer** - Go, MySQL, Security
- **UI/UX Designer** - User Interface & Experience
- **Security Engineer** - Application Security

## 📞 Support

Untuk pertanyaan atau dukungan teknis:
- 📧 Email: support@saferelief.id
- 🐛 Issues: [GitHub Issues](https://github.com/username/saferelief/issues)
- 📖 Documentation: [Wiki](https://github.com/username/saferelief/wiki)

---

**SafeRelief** - Menghubungkan kebaikan hati dengan mereka yang membutuhkan 💝