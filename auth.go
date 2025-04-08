package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID          int
	Name        string
	Email       string
	Password    string
	IsAdmin     bool
	DateCreated time.Time
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func comparePasswords(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func registerUser(name, email, password string) (int, error) {
	// Check if user already exists
	var id int
	err := db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err == nil {
		return 0, errors.New("user with this email already exists")
	} else if err != sql.ErrNoRows {
		return 0, err
	}

	// Hash the password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return 0, err
	}

	// Create new user
	result, err := db.Exec(
		"INSERT INTO users (name, email, password, is_admin) VALUES (?, ?, ?, ?)",
		name, email, hashedPassword, 0,
	)
	if err != nil {
		return 0, err
	}

	// Get the last inserted ID
	lastID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastID), nil
}

func loginUser(email, password string) (string, int, error) {
	// Find user
	var user User
	err := db.QueryRow(
		"SELECT id, password FROM users WHERE email = ?",
		email,
	).Scan(&user.ID, &user.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			return "", 0, errors.New("invalid email or password")
		}
		return "", 0, err
	}

	// Check password
	if !comparePasswords(user.Password, password) {
		return "", 0, errors.New("invalid email or password")
	}

	// Generate session token
	token, err := generateToken()
	if err != nil {
		return "", 0, err
	}

	// Create session
	expiresAt := time.Now().Add(24 * time.Hour) // 24-hour session
	_, err = db.Exec(
		"INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)",
		user.ID, token, expiresAt,
	)
	if err != nil {
		return "", 0, err
	}

	return token, user.ID, nil
}

func getUser(userID int) (User, error) {
	var user User
	err := db.QueryRow(
		"SELECT id, name, email, is_admin, date_created FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Name, &user.Email, &user.IsAdmin, &user.DateCreated)

	return user, err
}

func getUserFromSession(r *http.Request) (User, error) {
	sessionToken, err := r.Cookie("session")
	if err != nil {
		return User{}, err
	}

	var userID int
	err = db.QueryRow(
		"SELECT user_id FROM sessions WHERE token = ? AND expires_at > ?",
		sessionToken.Value, time.Now(),
	).Scan(&userID)

	if err != nil {
		return User{}, err
	}

	return getUser(userID)
}

func isAuthenticated(r *http.Request) bool {
	_, err := getUserFromSession(r)
	return err == nil
}

func isAdmin(r *http.Request) bool {
	user, err := getUserFromSession(r)
	if err != nil {
		return false
	}
	return user.IsAdmin
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func adminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !isAdmin(r) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
