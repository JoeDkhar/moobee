package main

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

func saveMovie(movie *Movie) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var result sql.Result
	if movie.ID == 0 {
		// Insert new movie
		result, err = tx.Exec(
			"INSERT INTO movies (title, time, duration, image, price) VALUES (?, ?, ?, ?, ?)",
			movie.Title, movie.Time, movie.Duration, movie.Image, movie.Price,
		)
		if err != nil {
			return err
		}

		// Get the inserted ID
		lastID, err := result.LastInsertId()
		if err != nil {
			return err
		}
		movie.ID = int(lastID)

		// Initialize empty seat grid
		for r := 0; r < 8; r++ {
			for c := 0; c < 10; c++ {
				_, err = tx.Exec(
					"INSERT INTO seats (movie_id, row, col, is_booked) VALUES (?, ?, ?, ?)",
					movie.ID, r, c, 0,
				)
				if err != nil {
					return err
				}
			}
		}

		// After getting the ID for a new movie
		if movie.ID > 0 && movie.Image == "" {
			movie.Image = GetImageURL(movie.ID)
			// Update the image in the database
			_, err = tx.Exec("UPDATE movies SET image = ? WHERE id = ?", movie.Image, movie.ID)
			if err != nil {
				log.Printf("Error updating image path: %v", err)
			}
		}
	} else {
		// Update existing movie
		_, err = tx.Exec(
			"UPDATE movies SET title = ?, time = ?, duration = ?, image = ?, price = ? WHERE id = ?",
			movie.Title, movie.Time, movie.Duration, movie.Image, movie.Price, movie.ID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func deleteMovie(id int) error {
	// SQLite with cascade will handle deleting associated records
	_, err := db.Exec("DELETE FROM movies WHERE id = ?", id)
	return err
}

func adminMovieHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Parse form data
		err := r.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		// Extract movie data
		idStr := r.FormValue("id")
		title := r.FormValue("title")
		showtime := r.FormValue("time")
		duration := r.FormValue("duration")

		price, err := strconv.ParseFloat(r.FormValue("price"), 64)
		if err != nil {
			price = 0
		}

		var id int
		if idStr != "" {
			id, _ = strconv.Atoi(idStr)
		}

		// Create or update movie
		movie := &Movie{
			ID:       id,
			Title:    title,
			Time:     showtime,
			Duration: duration,
			Price:    price,
		}

		// Handle image upload
		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()

			// Save the uploaded file
			imagePath := GetImagePath(movie.ID)
			if movie.ID == 0 {
				// For new movies, we'll set the image after we get the ID
				movie.Image = ""
			} else {
				movie.Image = GetImageURL(movie.ID)

				// Save the file
				f, err := os.Create(imagePath)
				if err != nil {
					http.Error(w, "Error saving image: "+err.Error(), http.StatusInternalServerError)
					return
				}
				defer f.Close()
				io.Copy(f, file)
			}
		} else {
			// No new image uploaded
			if movie.ID > 0 {
				// Keep existing image for updates
				var existingImage string
				err := db.QueryRow("SELECT image FROM movies WHERE id = ?", movie.ID).Scan(&existingImage)
				if err == nil {
					movie.Image = existingImage
				} else {
					movie.Image = GetImageURL(movie.ID)
				}
			} else {
				// Use default image for new movies
				movie.Image = "/static/images/default.jpg"
			}
		}

		// Save the movie
		err = saveMovie(movie)
		if err != nil {
			http.Error(w, "Error saving movie: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// For new movies, now that we have the ID, we can save the image properly
		if movie.ID > 0 && movie.Image == "" {
			movie.Image = GetImageURL(movie.ID)
			_, err = db.Exec("UPDATE movies SET image = ? WHERE id = ?", movie.Image, movie.ID)
			if err != nil {
				log.Printf("Error updating image path: %v", err)
			}

			// If we had a file uploaded, save it with the proper name
			if file != nil {
				// Seek back to beginning of file
				file.Seek(0, 0)

				// Save with proper name
				imagePath := GetImagePath(movie.ID)
				f, err := os.Create(imagePath)
				if err == nil {
					defer f.Close()
					io.Copy(f, file)
				}
			}
		}

		// Refresh the movies slice
		loadMovies()

		http.Redirect(w, r, "/admin/movies", http.StatusSeeOther)
		return
	}

	// Display the form with movie list
	err := templates.ExecuteTemplate(w, "admin_movies", movies)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func adminDeleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Path[len("/admin/movies/delete/"):]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	err = deleteMovie(id)
	if err != nil {
		http.Error(w, "Error deleting movie: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Refresh the movies slice
	loadMovies()

	http.Redirect(w, r, "/admin/movies", http.StatusSeeOther)
}

// Fix the end of loadMovies function around line 272
func loadMovies() {
	// Load movies from SQLite
	rows, err := db.Query("SELECT id, title, time, duration, image, price FROM movies")
	if err != nil {
		log.Println("Error loading movies:", err)
		return
	}
	defer rows.Close()

	var loadedMovies []Movie
	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.Time, &movie.Duration, &movie.Image, &movie.Price); err != nil {
			log.Println("Error scanning movie row:", err)
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
			log.Println("Error loading seats for movie", movie.ID, ":", err)
			continue
		}

		for seatRows.Next() {
			var row, col, isBooked int
			if err := seatRows.Scan(&row, &col, &isBooked); err != nil {
				log.Println("Error scanning seat row:", err)
				continue
			}

			if row < len(movie.Seats) && col < len(movie.Seats[row]) {
				movie.Seats[row][col] = isBooked == 1
			}
		}
		seatRows.Close()

		loadedMovies = append(loadedMovies, movie)
	}

	// Replace the global movies slice
	mutex.Lock()
	movies = loadedMovies
	mutex.Unlock()

	// Download images without updating the database directly

	// If no movies were found, initialize with sample data
	// IMPORTANT: Notice the change here to avoid recursion
	if len(movies) == 0 {
		createSampleMovies() // Use a new function that doesn't call loadMovies
	}
}

// Add this new function that doesn't call loadMovies
func createSampleMovies() {
	// Create sample movies
	sampleMovies := []Movie{
		{
			Title:    "Spider-Man: No Way Home",
			Time:     "2023-08-01 18:00",
			Duration: "2h 28m",
			Price:    14.99,
			Image:    "/static/images/movie_1.jpg",
		},
		{
			Title:    "Dead Poets Society",
			Time:     "2023-08-02 16:30",
			Duration: "2h 8m",
			Price:    11.99,
			Image:    "/static/images/movie_2.jpg",
		},
		{
			Title:    "The Shawshank Redemption",
			Time:     "2023-08-03 19:15",
			Duration: "2h 22m",
			Price:    12.99,
			Image:    "/static/images/movie_3.jpg",
		},
		{
			Title:    "Inception",
			Time:     "2023-08-04 20:30",
			Duration: "2h 28m",
			Price:    13.99,
			Image:    "/static/images/movie_4.jpg",
		},
		{
			Title:    "The Matrix",
			Time:     "2023-08-01 21:15",
			Duration: "2h 16m",
			Price:    12.99,
			Image:    "/static/images/movie_5.jpg",
		},
		{
			Title:    "Interstellar",
			Time:     "2023-08-02 19:00",
			Duration: "2h 49m",
			Price:    15.99,
			Image:    "/static/images/movie_6.jpg",
		},
		{
			Title:    "Pulp Fiction",
			Time:     "2023-08-03 20:00",
			Duration: "2h 34m",
			Price:    13.50,
			Image:    "/static/images/movie_7.jpg",
		},
		{
			Title:    "The Dark Knight",
			Time:     "2023-08-04 18:45",
			Duration: "2h 32m",
			Price:    14.50,
			Image:    "/static/images/movie_8.jpg",
		},
		{
			Title:    "Parasite",
			Time:     "2023-08-05 17:30",
			Duration: "2h 12m",
			Price:    13.99,
			Image:    "/static/images/movie_9.jpg",
		},
	}

	// Initialize seats for each movie (8 rows x 10 columns)
	for i := range sampleMovies {
		sampleMovies[i].Seats = make([][]bool, 8)
		for j := range sampleMovies[i].Seats {
			sampleMovies[i].Seats[j] = make([]bool, 10)
		}
	}

	// Save sample movies to database
	for i, movie := range sampleMovies {
		movie.ID = i + 1 // Ensure sequential IDs
		movie.Image = fmt.Sprintf("/static/images/movie_%d.jpg", movie.ID)
		if err := saveMovie(&movie); err != nil {
			log.Println("Error saving sample movie:", err)
		}
	}

	// Update the movies slice directly
	mutex.Lock()
	movies = sampleMovies
	mutex.Unlock()
}

// Change initSampleMovies to use the new function
func initSampleMovies() {
	createSampleMovies()
	// DO NOT call loadMovies() here
}
