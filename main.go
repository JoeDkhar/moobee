package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	movies      []Movie
	mutex       sync.Mutex
	templates   *template.Template
	cssTemplate *template.Template // Add this line
)

type Movie struct {
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Time     string   `json:"time"`
	Duration string   `json:"duration"`
	Image    string   `json:"image"`
	Price    float64  `json:"price"`
	Seats    [][]bool `json:"-"` // Not stored in DB directly, loaded separately
}

type Booking struct {
	ID      int
	UserID  int
	Name    string
	Email   string
	MovieID int
	Seats   []string
	Total   float64
	Date    time.Time
}

type BookingResponse struct {
	Success   bool
	Message   string
	BookingID int
}

type BookingRequest struct {
	Name    string   `json:"name"`
	Email   string   `json:"email"`
	MovieID int      `json:"movieID"`
	Seats   []string `json:"seats"`
}

func main() {
	// Initialize database
	err := InitDB()
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer db.Close()

	// Create static directories
	os.MkdirAll("static/images", 0755)

	// Load movies (call the function from admin.go)
	loadMovies()

	// Initialize templates
	initTemplates()

	// Setup routes for static files and handlers
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Setup API routes
	http.HandleFunc("/api/book", apiBookHandler)

	// Setup page routes
	http.HandleFunc("/", landingHandler)  // Landing page is now the root
	http.HandleFunc("/home", homeHandler) // Home page moved to /home
	http.HandleFunc("/book/", bookHandler)
	http.HandleFunc("/booking/", viewBookingHandler)
	http.HandleFunc("/bookings", bookingsHandler)
	http.HandleFunc("/cancel/", cancelBookingHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/search", searchHandler)
	http.HandleFunc("/admin", adminHandler)
	http.HandleFunc("/admin/movies", adminMovieHandler)
	http.HandleFunc("/admin/movies/delete/", adminDeleteMovieHandler)

	// Also register the CSS handler
	http.HandleFunc("/static/styles.css", staticHandler)

	// Start the server
	log.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initTemplates() {
	// Create new template with custom functions
	templates = template.New("").Funcs(template.FuncMap{
		"availableSeats": availableSeats,
		"formatPrice":    formatPrice,
		"add":            func(a, b int) int { return a + b },
		"getMovie":       getMovie,
	})

	// Parse all templates
	templates.New("landing").Parse(landingTemplate) // Add this line
	templates.New("home").Parse(homeTemplate)
	templates.New("book").Parse(bookTemplate)
	templates.New("bookings").Parse(bookingsTemplate)
	templates.New("view_booking").Parse(viewBookingTemplate)
	templates.New("login").Parse(loginTemplate)
	templates.New("register").Parse(registerTemplate)
	templates.New("profile").Parse(profileTemplate)
	templates.New("admin").Parse(adminTemplate)
	templates.New("admin_movies").Parse(adminMoviesTemplate)
	templates.New("search").Parse(searchTemplate)

	// This is for the static css handler
	cssTemplate = template.Must(template.New("css").Parse(cssContent))
}

func bookHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/book/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	var movie *Movie
	for i := range movies {
		if movies[i].ID == id {
			movie = &movies[i]
			break
		}
	}

	if movie == nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	// Get user if logged in
	user, _ := getUserFromSession(r)

	data := struct {
		Movie *Movie
		User  User
	}{
		Movie: movie,
		User:  user,
	}

	err = templates.ExecuteTemplate(w, "book", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func apiBookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Parse JSON request body
	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONResponse(w, BookingResponse{Success: false, Message: "Invalid request data"})
		return
	}

	if req.Name == "" || req.Email == "" || len(req.Seats) == 0 {
		sendJSONResponse(w, BookingResponse{Success: false, Message: "Missing required fields"})
		return
	}

	// Find the movie
	var movie *Movie
	for i := range movies {
		if movies[i].ID == req.MovieID {
			movie = &movies[i]
			break
		}
	}

	if movie == nil {
		sendJSONResponse(w, BookingResponse{Success: false, Message: "Movie not found"})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	// Check seat availability
	for _, seatStr := range req.Seats {
		var row, col int
		fmt.Sscanf(seatStr, "%d-%d", &row, &col)
		if row >= len(movie.Seats) || col >= len(movie.Seats[row]) || movie.Seats[row][col] {
			sendJSONResponse(w, BookingResponse{Success: false, Message: fmt.Sprintf("Seat %s is not available", seatStr)})
			return
		}
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		sendJSONResponse(w, BookingResponse{Success: false, Message: "Database error"})
		return
	}
	defer tx.Rollback()

	// Get user ID if logged in
	var userID int
	user, err := getUserFromSession(r)
	if err == nil {
		userID = user.ID
	}

	// Calculate total price
	total := float64(len(req.Seats)) * movie.Price

	// Create booking
	result, err := tx.Exec(
		"INSERT INTO bookings (user_id, name, email, movie_id, total) VALUES (?, ?, ?, ?, ?)",
		userID, req.Name, req.Email, movie.ID, total,
	)
	if err != nil {
		sendJSONResponse(w, BookingResponse{Success: false, Message: "Error creating booking"})
		return
	}

	// Get the booking ID
	bookingID, err := result.LastInsertId()
	if err != nil {
		sendJSONResponse(w, BookingResponse{Success: false, Message: "Error getting booking ID"})
		return
	}

	// Save the seats
	for _, seatStr := range req.Seats {
		var row, col int
		fmt.Sscanf(seatStr, "%d-%d", &row, &col)

		// Update the seats table
		_, err = tx.Exec(
			"UPDATE seats SET is_booked = 1 WHERE movie_id = ? AND row = ? AND col = ?",
			movie.ID, row, col,
		)
		if err != nil {
			sendJSONResponse(w, BookingResponse{Success: false, Message: "Error updating seat"})
			return
		}

		// Insert into booking_seats
		_, err = tx.Exec(
			"INSERT INTO booking_seats (booking_id, row, col) VALUES (?, ?, ?)",
			bookingID, row, col,
		)
		if err != nil {
			sendJSONResponse(w, BookingResponse{Success: false, Message: "Error saving seat selection"})
			return
		}

		// Update in-memory seats
		movie.Seats[row][col] = true
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		sendJSONResponse(w, BookingResponse{Success: false, Message: "Error finalizing booking"})
		return
	}

	sendJSONResponse(w, BookingResponse{
		Success:   true,
		Message:   "Booking successful",
		BookingID: int(bookingID),
	})
}

func sendJSONResponse(w http.ResponseWriter, response BookingResponse) {
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error creating JSON response", http.StatusInternalServerError)
		return
	}

	w.Write(jsonResponse)
}

func bookingsHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var bookings []Booking
	var rows *sql.Rows

	if user.IsAdmin {
		// Admins see all bookings
		rows, err = db.Query(`
            SELECT b.id, b.user_id, b.name, b.email, b.movie_id, b.total, b.date 
            FROM bookings b 
            ORDER BY b.date DESC
        `)
	} else {
		// Regular users see only their bookings
		rows, err = db.Query(`
            SELECT b.id, b.user_id, b.name, b.email, b.movie_id, b.total, b.date 
            FROM bookings b 
            WHERE b.user_id = ? OR b.email = ?
            ORDER BY b.date DESC
        `, user.ID, user.Email)
	}

	if err != nil {
		http.Error(w, "Error loading bookings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Scan the rows
	for rows.Next() {
		var b Booking
		var userID sql.NullInt64
		if err := rows.Scan(&b.ID, &userID, &b.Name, &b.Email, &b.MovieID, &b.Total, &b.Date); err != nil {
			log.Printf("Error scanning booking row: %v", err)
			continue
		}

		if userID.Valid {
			b.UserID = int(userID.Int64)
		}

		// Get the seats for this booking
		b.Seats = getBookingSeats(b.ID)
		bookings = append(bookings, b)
	}

	data := struct {
		Bookings []Booking
		User     User
	}{
		Bookings: bookings,
		User:     user,
	}

	if err := templates.ExecuteTemplate(w, "bookings", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func getBookingSeats(bookingID int) []string {
	rows, err := db.Query(
		"SELECT row, col FROM booking_seats WHERE booking_id = ?",
		bookingID,
	)
	if err != nil {
		log.Printf("Error getting seats for booking %d: %v", bookingID, err)
		return nil
	}
	defer rows.Close()

	var seats []string
	for rows.Next() {
		var row, col int
		if err := rows.Scan(&row, &col); err != nil {
			log.Printf("Error scanning seat row: %v", err)
			continue
		}
		seats = append(seats, fmt.Sprintf("%d-%d", row, col))
	}

	return seats
}

func viewBookingHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/booking/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	// Get the booking
	var booking Booking
	var userID sql.NullInt64
	err = db.QueryRow(`
        SELECT id, user_id, name, email, movie_id, total, date 
        FROM bookings 
        WHERE id = ?
    `, id).Scan(&booking.ID, &userID, &booking.Name, &booking.Email, &booking.MovieID, &booking.Total, &booking.Date)

	if err != nil {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	if userID.Valid {
		booking.UserID = int(userID.Int64)
	}

	// Get the seats
	booking.Seats = getBookingSeats(booking.ID)

	// Get the user if logged in
	user, _ := getUserFromSession(r)

	// Check if the user is authorized to view this booking
	if !user.IsAdmin && user.ID != booking.UserID && user.Email != booking.Email {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	data := struct {
		Booking Booking
		User    User
	}{
		Booking: booking,
		User:    user,
	}

	err = templates.ExecuteTemplate(w, "view_booking", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func cancelBookingHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/cancel/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	// Get the current user
	user, err := getUserFromSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Get the booking
	var booking Booking
	var userID sql.NullInt64
	err = tx.QueryRow(`
        SELECT id, user_id, email, movie_id
        FROM bookings 
        WHERE id = ?
    `, id).Scan(&booking.ID, &userID, &booking.Email, &booking.MovieID)

	if err != nil {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	if userID.Valid {
		booking.UserID = int(userID.Int64)
	}

	// Check if the user is authorized to cancel this booking
	if !user.IsAdmin && user.ID != booking.UserID && user.Email != booking.Email {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the seats
	rows, err := tx.Query(
		"SELECT row, col FROM booking_seats WHERE booking_id = ?",
		id,
	)
	if err != nil {
		http.Error(w, "Error getting booking seats", http.StatusInternalServerError)
		return
	}

	var seatRows, seatCols []int
	for rows.Next() {
		var row, col int
		if err := rows.Scan(&row, &col); err != nil {
			rows.Close()
			http.Error(w, "Error reading seat data", http.StatusInternalServerError)
			return
		}
		seatRows = append(seatRows, row)
		seatCols = append(seatCols, col)
	}
	rows.Close()

	// Update seats to be available again
	for i := range seatRows {
		_, err = tx.Exec(
			"UPDATE seats SET is_booked = 0 WHERE movie_id = ? AND row = ? AND col = ?",
			booking.MovieID, seatRows[i], seatCols[i],
		)
		if err != nil {
			http.Error(w, "Error updating seat status", http.StatusInternalServerError)
			return
		}

		// Update in-memory seats
		for j := range movies {
			if movies[j].ID == booking.MovieID {
				if seatRows[i] < len(movies[j].Seats) && seatCols[i] < len(movies[j].Seats[seatRows[i]]) {
					movies[j].Seats[seatRows[i]][seatCols[i]] = false
				}
				break
			}
		}
	}

	// Delete the booking_seats entries
	_, err = tx.Exec("DELETE FROM booking_seats WHERE booking_id = ?", id)
	if err != nil {
		http.Error(w, "Error deleting seat reservations", http.StatusInternalServerError)
		return
	}

	// Delete the booking
	_, err = tx.Exec("DELETE FROM bookings WHERE id = ?", id)
	if err != nil {
		http.Error(w, "Error canceling booking", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error finalizing cancellation", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/bookings", http.StatusSeeOther)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/css")
	w.Write([]byte(cssContent))
}

func availableSeats(seats [][]bool) int {
	count := 0
	for _, row := range seats {
		for _, booked := range row {
			if !booked {
				count++
			}
		}
	}
	return count
}

func formatPrice(price float64) string {
	return fmt.Sprintf("$%.2f", price)
}

func getMovie(id int) *Movie {
	for i := range movies {
		if movies[i].ID == id {
			return &movies[i]
		}
	}
	return nil
}
