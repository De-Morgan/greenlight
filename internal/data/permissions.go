package data

import (
	"context"
	"database/sql"
	"slices"
	"time"

	"github.com/lib/pq"
)

type Permission string

const (
	PermissionMovieRead  Permission = "movies:read"
	PermissionMovieWrite Permission = "movies:write"
)

type Permissions []Permission

// Includes check whether the Permissions slice contains permission code
func (p Permissions) Includes(code Permission) bool {
	return slices.Contains(p, code)
}

type PermissionModel struct {
	DB *sql.DB
}

func (m PermissionModel) GetAllForUser(userId int64) (Permissions, error) {
	query :=
		`SELECT p.code
		FROM permissions p
		INNER JOIN users_permissions up ON up.permission_id = p.id
		INNER JOIN users u ON up.user_id = u.id
		WHERE u.id  = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions Permissions
	for rows.Next() {
		var permission Permission
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil

}

func (m PermissionModel) AddForUser(userId int64, permissions ...Permission) error {

	query := `
		INSERT INTO users_permissions
		SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, userId, pq.Array(permissions))
	return err
}
