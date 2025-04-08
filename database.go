package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

// InitDB initializes the SQLite database connection and creates tables if they don't exist
func InitDB() error {
	dataDir := "data"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		os.Mkdir(dataDir, 0755)
	}

	var err error
	db, err = sql.Open("sqlite3", "data/cinema.db")
	if err != nil {
		return err
	}

	// Test connection
	if err = db.Ping(); err != nil {
		return err
	}

	// Create tables if they don't exist
	if err = createTables(); err != nil {
		return err
	}

	// Create default admin user
	if err = ensureAdminUser(); err != nil {
		log.Printf("Warning: Failed to create admin user: %v", err)
	}

	return nil
}

// createTables creates all required tables in the database
func createTables() error {
	// Create movies table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS movies (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			time TEXT NOT NULL,
			duration TEXT NOT NULL,
			image TEXT,
			price REAL NOT NULL
        )
    `)
	if err != nil {
		return err
	}

	// Create seats table
	_, err = db.Exec(`
         CREATE TABLE IF NOT EXISTS seats (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            movie_id INTEGER NOT NULL,
            row INTEGER NOT NULL,
            col INTEGER NOT NULL,
            is_booked INTEGER NOT NULL DEFAULT 0,
            FOREIGN KEY (movie_id) REFERENCES movies (id) ON DELETE CASCADE
        )
    `)
	if err != nil {
		return err
	}

	// Create users table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL,
            email TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL,
            is_admin INTEGER NOT NULL DEFAULT 0,
            date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
	if err != nil {
		return err
	}

	// Create bookings table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS bookings (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER,
            movie_id INTEGER NOT NULL,
            name TEXT NOT NULL,
            email TEXT NOT NULL,
            total REAL NOT NULL,
            date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL,
            FOREIGN KEY (movie_id) REFERENCES movies (id) ON DELETE CASCADE
        )
    `)
	if err != nil {
		return err
	}

	// Create booking_seats table
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS booking_seats (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            booking_id INTEGER NOT NULL,
            row INTEGER NOT NULL,
            col INTEGER NOT NULL,
            FOREIGN KEY (booking_id) REFERENCES bookings (id) ON DELETE CASCADE
        )
    `)
	if err != nil {
		return err
	}

	// Add this new table for sessions
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS sessions (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            user_id INTEGER NOT NULL,
            token TEXT NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
        )
    `)

	return err
}

// ensureAdminUser makes sure there's at least one admin user in the system
func ensureAdminUser() error {
	// Check if admin exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE is_admin = 1").Scan(&count)
	if err != nil {
		return err
	}

	// If no admin, create one
	if count == 0 {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		_, err = db.Exec(
			"INSERT INTO users (name, email, password, is_admin) VALUES (?, ?, ?, ?)",
			"Admin", "admin@example.com", string(hashedPassword), 1,
		)
		if err != nil {
			return err
		}

		log.Println("Created admin user: admin@example.com / admin123")
	}

	return nil
}

// CreateImagesDirectory creates the directory for storing movie images
func CreateImagesDirectory() error {
	imagesDir := filepath.Join("static", "images")
	return os.MkdirAll(imagesDir, 0755)
}

// GetImagePath returns the local path for a movie image
func GetImagePath(movieID int) string {
	return filepath.Join("static", "images", fmt.Sprintf("movie_%d.jpg", movieID))
}

// GetImageURL returns the URL for a movie image
func GetImageURL(movieID int) string {
	return fmt.Sprintf("/static/images/movie_%d.jpg", movieID)
}
