package storage

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"frame_control_system/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u models.User) error {
	now := time.Now().UTC().Format(time.RFC3339)
	roles := strings.Join(u.Roles, ",")
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, name, roles, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, u.ID, u.Email, u.PasswordHash, u.Name, roles, now, now)
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, name, roles, created_at, updated_at
		FROM users WHERE email = ?
	`, email)
	var (
		u     models.User
		roles string
		createdAt, updatedAt string
	)
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &roles, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	u.Roles = splitRoles(roles)
	u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &u, nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	row := r.db.QueryRowContext(ctx, `SELECT 1 FROM users WHERE email = ? LIMIT 1`)
	var one int
	if err := row.Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func splitRoles(s string) []string {
	if s == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}


