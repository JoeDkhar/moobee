package main

import (
	"fmt"
	"path/filepath"
)

// Returns the file system path for a movie image
func GetImagePathHelper(movieID int) string {
	return filepath.Join("static", "images", fmt.Sprintf("movie_%d.jpg", movieID))
}

// Returns the URL for accessing a movie image
func GetImageURLHelper(movieID int) string {
	return fmt.Sprintf("/static/images/movie_%d.jpg", movieID)
}
