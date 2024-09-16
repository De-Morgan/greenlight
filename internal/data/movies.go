package data

import (
	"strings"
	"time"

	"morgan.greenlight.nex/internal/validator"
)

type Movie struct {
	ID        int64     `json:"id"`                // Unique integer ID for the movie
	CreatedAt time.Time `json:"-"`                 // Timestamp for when the movie is added to our database
	Title     string    `json:"title"`             // Movie  title
	Year      int32     `json:"year,omitempty"`    // Movie release year
	Runtime   Runtime   `json:"runtime,omitempty"` // Movie runtime (in minutes)
	Genres    []string  `json:"genres,omitempty"`  // Slice of genres for the movie (romance, comedy, etc.)
	Version   int32     `json:"version"`           // The version number starts at 1 and will be incremented each time the movie information is updated
}

func (m Movie) Validate() (bool, map[string]string) {
	v := validator.New()
	v.Check(strings.TrimSpace(m.Title) == "", "title", "Title cannot be empty")
	v.Check(len(m.Title) > 500, "title", "must not be more than 500 bytes long")
	v.Check(m.Year == 0, "year", "year must be provided")
	v.Check(m.Year <= 1888, "year", "year must be greater than 1888")
	v.Check(m.Runtime <= 0, "runtime", "Runtime must be positive integer greater than zero")
	v.Check(len(m.Genres) > 5 || len(m.Genres) < 1, "genres", "Genres must be between 1 and 5 items")
	v.Check(!validator.Unique(m.Genres), "genres", "Genres must not contain duplicate values")
	return v.Valid(), v.Errors
}
