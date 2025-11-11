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

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, name, roles, created_at, updated_at
		FROM users WHERE id = ?
	`, id)
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

func (r *UserRepository) UpdateName(ctx context.Context, id, name string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		UPDATE users SET name = ?, updated_at = ? WHERE id = ?
	`, name, now, id)
	return err
}

type ListUsersParams struct {
	Email string
	Name  string
	Role  string
	Limit int
	Offset int
	Sort string
}

func (r *UserRepository) List(ctx context.Context, p ListUsersParams) ([]models.User, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if p.Email != "" {
		where = append(where, "email LIKE ?")
		args = append(args, "%"+p.Email+"%")
	}
	if p.Name != "" {
		where = append(where, "name LIKE ?")
		args = append(args, "%"+p.Name+"%")
	}
	if p.Role != "" {
		where = append(where, "roles LIKE ?")
		args = append(args, "%"+p.Role+"%")
	}
	order := "created_at DESC"
	switch p.Sort {
	case "email_asc":
		order = "email ASC"
	case "email_desc":
		order = "email DESC"
	case "name_asc":
		order = "name ASC"
	case "name_desc":
		order = "name DESC"
	case "created_asc":
		order = "created_at ASC"
	case "created_desc":
		order = "created_at DESC"
	}
	query := `
		SELECT id, email, password_hash, name, roles, created_at, updated_at
		FROM users
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY ` + order + `
		LIMIT ? OFFSET ?
	`
	args = append(args, p.Limit, p.Offset)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []models.User
	for rows.Next() {
		var u models.User
		var roles, createdAt, updatedAt string
		if err := rows.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &roles, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		u.Roles = splitRoles(roles)
		u.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		u.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		res = append(res, u)
	}
	return res, rows.Err()
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


