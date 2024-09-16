package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

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
