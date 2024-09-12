package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
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

type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
		INSERT INTO movies (title,year,runtime,genres)
		VALUES($1, $2, $3, $4)
		RETURNING id, created_at, version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{
		movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres),
	}

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		WHERE id = $1
	`
	var movie Movie

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query :=
		`UPDATE movies SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
	 	 WHERE id = $5 AND version = $6
	 	 RETURNING version`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres), movie.ID, movie.Version}
	if err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
		DELETE FROM movies
		WHERE id = $1
	`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
func (m MovieModel) GetAll(year int32, genres []string, filters Filters) ([]*Movie, Metadata, error) {

	var query string

	if filters.SortPresent() {
		query = fmt.Sprintf(` SELECT  count(*) OVER(), id, created_at, title, year, runtime, genres, version
		 FROM movies 
		 WHERE (year = $1 OR $1 = 0)
		 AND (genres @> $2 OR $2 = '{}')
		 AND (to_tsvector('simple',title) @@  plainto_tsquery('simple',$3) OR $3 = '')
		 ORDER BY %s %s, id ASC
		 LIMIT $4 OFFSET $5`, filters.SortColumn(), filters.SortDirection())
	} else {
		query = ` SELECT  count(*) OVER(), id, created_at, title, year, runtime, genres, version
		 FROM movies 
		 WHERE (year = $1 OR $1 = 0)
		 AND (genres @> $2 OR $2 = '{}')
		 AND (to_tsvector('simple',title) @@  plainto_tsquery('simple',$3) OR $3 = '')
		 ORDER BY id ASC
		 LIMIT $4 OFFSET $5`
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, year, pq.Array(genres), filters.SearchTerm, filters.Limit, filters.OffSet())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	movies := []*Movie{}

	for rows.Next() {
		var movie Movie
		err := rows.Scan(&totalRecords, &movie.ID, &movie.CreatedAt, &movie.Title, &movie.Year, &movie.Runtime, pq.Array(&movie.Genres), &movie.Version)
		if err != nil {
			return nil, Metadata{}, err
		}
		movies = append(movies, &movie)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.Limit)
	return movies, metadata, nil

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
