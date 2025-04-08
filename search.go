package main

import (
	"net/http"
)

func searchMovies(query string) []Movie {
	if query == "" {
		return movies
	}

	// Case-insensitive search in SQLite
	rows, err := db.Query(
		"SELECT id, title, time, duration, image, price FROM movies WHERE title LIKE ?",
		"%"+query+"%",
	)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []Movie
	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Time, &movie.Duration, &movie.Image, &movie.Price); err != nil {
			continue
		}

		// Initialize the seats array
		movie.Seats = make([][]bool, 8)
		for i := range movie.Seats {
			movie.Seats[i] = make([]bool, 10)
		}

		// Load seat information
		seatRows, err := db.Query("SELECT row, col, is_booked FROM seats WHERE movie_id = ?", movie.ID)
		if err != nil {
			continue
		}

		for seatRows.Next() {
			var row, col, isBooked int
			if err := seatRows.Scan(&row, &col, &isBooked); err != nil {
				continue
			}

			if row < len(movie.Seats) && col < len(movie.Seats[row]) {
				movie.Seats[row][col] = isBooked == 1
			}
		}
		seatRows.Close()

		results = append(results, movie)
	}

	return results
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	results := searchMovies(query)

	// Get user if logged in
	user, _ := getUserFromSession(r)

	data := struct {
		Query  string
		Movies []Movie
		User   User
	}{
		Query:  query,
		Movies: results,
		User:   user,
	}

	err := templates.ExecuteTemplate(w, "search", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
