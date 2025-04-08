package main

import (
	"log"
	"net/http"
	"time"
)

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := struct {
		Error string
	}{}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		token, _, err := loginUser(email, password)
		if err != nil {
			data.Error = err.Error()
		} else {
			// Set session cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    token,
				Path:     "/",
				Expires:  time.Now().Add(24 * time.Hour),
				HttpOnly: true,
			})

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}

	err := templates.ExecuteTemplate(w, "login", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear the session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if isAuthenticated(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := struct {
		Error string
	}{}

	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		password2 := r.FormValue("password2")

		if password != password2 {
			data.Error = "Passwords do not match"
		} else if name == "" || email == "" || password == "" {
			data.Error = "All fields are required"
		} else {
			_, err := registerUser(name, email, password)
			if err != nil {
				data.Error = err.Error()
			} else {
				// Automatically log in the user
				token, _, err := loginUser(email, password)
				if err != nil {
					data.Error = "Registration successful, but could not log in: " + err.Error()
				} else {
					// Set session cookie
					http.SetCookie(w, &http.Cookie{
						Name:     "session",
						Value:    token,
						Path:     "/",
						Expires:  time.Now().Add(24 * time.Hour),
						HttpOnly: true,
					})

					http.Redirect(w, r, "/", http.StatusSeeOther)
					return
				}
			}
		}
	}

	err := templates.ExecuteTemplate(w, "register", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get user's bookings
	rows, err := db.Query(`
        SELECT id, name, email, movie_id, total, date
        FROM bookings 
        WHERE user_id = ? OR email = ?
        ORDER BY date DESC
        LIMIT 5
    `, user.ID, user.Email)

	if err != nil {
		http.Error(w, "Error loading bookings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.Name, &b.Email, &b.MovieID, &b.Total, &b.Date); err != nil {
			log.Printf("Error scanning booking row: %v", err)
			continue
		}

		// Get the seats for this booking
		b.Seats = getBookingSeats(b.ID)
		bookings = append(bookings, b)
	}

	data := struct {
		User     User
		Bookings []Booking
	}{
		User:     user,
		Bookings: bookings,
	}

	err = templates.ExecuteTemplate(w, "profile", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func adminHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromSession(r)
	if err != nil || !user.IsAdmin {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get statistics for admin dashboard
	var movieCount, bookingCount, userCount int
	var totalRevenue float64

	// Count movies
	err = db.QueryRow("SELECT COUNT(*) FROM movies").Scan(&movieCount)
	if err != nil {
		movieCount = 0
	}

	// Count bookings
	err = db.QueryRow("SELECT COUNT(*) FROM bookings").Scan(&bookingCount)
	if err != nil {
		bookingCount = 0
	}

	// Count users
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		userCount = 0
	}

	// Calculate total revenue
	err = db.QueryRow("SELECT COALESCE(SUM(total), 0) FROM bookings").Scan(&totalRevenue)
	if err != nil {
		totalRevenue = 0
	}

	// Get recent bookings
	rows, err := db.Query(`
        SELECT b.id, b.name, b.email, b.movie_id, b.total, b.date
        FROM bookings b
        ORDER BY b.date DESC 
        LIMIT 10
    `)
	if err != nil {
		http.Error(w, "Error loading recent bookings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recentBookings []Booking
	for rows.Next() {
		var b Booking
		if err := rows.Scan(&b.ID, &b.Name, &b.Email, &b.MovieID, &b.Total, &b.Date); err != nil {
			log.Printf("Error scanning booking row: %v", err)
			continue
		}

		// Get the seats for this booking
		b.Seats = getBookingSeats(b.ID)
		recentBookings = append(recentBookings, b)
	}

	data := struct {
		MovieCount     int
		BookingCount   int
		UserCount      int
		TotalRevenue   float64
		RecentBookings []Booking
		User           User
	}{
		MovieCount:     movieCount,
		BookingCount:   bookingCount,
		UserCount:      userCount,
		TotalRevenue:   totalRevenue,
		RecentBookings: recentBookings,
		User:           user,
	}

	err = templates.ExecuteTemplate(w, "admin", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Landing page handler
func landingHandler(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if isAuthenticated(r) {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}

	err := templates.ExecuteTemplate(w, "landing", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Home page handler (modified)
func homeHandler(w http.ResponseWriter, r *http.Request) {
	// If the request is for the root path only
	if r.URL.Path != "/" && r.URL.Path != "/home" {
		http.NotFound(w, r)
		return
	}

	// Get user from session
	user, err := getUserFromSession(r)

	// Create data for template
	data := struct {
		Movies []Movie
		User   User
	}{
		Movies: movies,
		User:   user,
	}

	// Execute template
	err = templates.ExecuteTemplate(w, "home", data)
	if err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
	}
}
