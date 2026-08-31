package main

import (
	"context"
	"flag"
	"os"
	"time"

	"devclub.com/identity/internal/api/config"
	"devclub.com/identity/internal/database"
	"devclub.com/identity/pkg/utils"
)

func main() {
	infoLogger, _, errorLogger := utils.CreateLoggers()

	cfg, err := config.NewConfig()
	if err != nil {
		errorLogger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	emailFlag := flag.String("email", os.Getenv("ADMIN_EMAIL"), "Email address for the initial admin user")
	passwordFlag := flag.String("password", os.Getenv("ADMIN_PASSWORD"), "Password for the initial admin user")
	flag.Parse()

	email := *emailFlag
	if email == "" {
		email = "admin@devclub.com"
	}

	password := *passwordFlag
	if password == "" {
		password = "Admin@123456"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DBConn, errorLogger)
	if err != nil {
		errorLogger.Error("failed to connect to database for seeding", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Check if admin user already exists (by email or role = 'admin')
	var existingID string
	checkQuery := `SELECT id FROM users WHERE email = $1 OR role = 'admin' LIMIT 1`
	err = pool.QueryRow(ctx, checkQuery, email).Scan(&existingID)
	if err == nil {
		infoLogger.Info("admin user already exists, skipping seed", "email", email, "user_id", existingID)
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		errorLogger.Error("failed to hash admin password", "error", err)
		os.Exit(1)
	}

	// Insert Admin User
	insertQuery := `
		INSERT INTO users (email, password_hash, role, is_active)
		VALUES ($1, $2, 'admin', TRUE)
		RETURNING id, email, role, created_at
	`

	var adminID, adminEmail, adminRole string
	var createdAt time.Time

	err = pool.QueryRow(ctx, insertQuery, email, hashedPassword).Scan(&adminID, &adminEmail, &adminRole, &createdAt)
	if err != nil {
		errorLogger.Error("failed to create admin user", "error", err)
		os.Exit(1)
	}

	infoLogger.Info("initial admin user seeded successfully!",
		"user_id", adminID,
		"email", adminEmail,
		"role", adminRole,
		"created_at", createdAt,
	)
}
