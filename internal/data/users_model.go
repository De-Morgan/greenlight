package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

// Define a custom ErrDuplicateEmail error.
var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) Insert(ctx context.Context, user *User) error {

	stmt := `
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, activated, version`

	args := []any{user.Name, user.Email, user.Password.hash}

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAt, &user.Activated, &user.Version)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "users_email_key"):
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(ctx context.Context, email string) (*User, error) {

	stmt := `
		SELECT  id, created_at, name, email, password_hash, activated, version FROM users
		WHERE email = $1`
	var user User
	err := m.DB.QueryRowContext(ctx, stmt, email).Scan(&user.ID, &user.CreatedAt, &user.Name, &user.Email, &user.Password.hash, &user.Activated, &user.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) UpdateUser(ctx context.Context, user *User) error {
	stmt := `
		UPDATE users
		SET name = $1, email = $2, password_hash = $3, activated = $4, version = version +1
		WHERE id = $5 AND version = $6
		RETURNING version`

	args := []any{&user.Name, &user.Email, &user.Password.hash, &user.Activated, &user.ID, &user.Version}

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.Version)

	if err != nil {
		switch {
		case strings.Contains(err.Error(), "users_email_key"):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(ctx context.Context, tokenscope TokenScope, tokenPlaintext string) (*User, error) {

	tokenHash := hashToken(tokenPlaintext)

	stmt :=
		`SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
		FROM users
		INNER JOIN tokens
		ON tokens.user_id=users.id
		WHERE tokens.hash=$1
		AND tokens.scope=$2
		AND tokens.expiry>$3`
	args := []any{string(tokenHash), string(tokenscope), time.Now()}

	var user User

	err := m.DB.QueryRowContext(ctx, stmt, args...).Scan(&user.ID, &user.CreatedAt, &user.Name, &user.Email, &user.Password.hash, &user.Activated, &user.Version)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}
