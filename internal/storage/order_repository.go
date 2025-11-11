package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"frame_control_system/internal/models"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o models.Order) error {
	itemsJSON, _ := json.Marshal(o.Items)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO orders (id, user_id, items, status, total_amount, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, o.ID, o.UserID, string(itemsJSON), string(o.Status), o.TotalAmount, now, now)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, items, status, total_amount, created_at, updated_at
		FROM orders WHERE id = ?
	`, id)
	var itemsStr, status, createdAt, updatedAt string
	var o models.Order
	if err := row.Scan(&o.ID, &o.UserID, &itemsStr, &status, &o.TotalAmount, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	_ = json.Unmarshal([]byte(itemsStr), &o.Items)
	o.Status = models.OrderStatus(status)
	o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	o.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &o, nil
}

type ListOrdersParams struct {
	UserID string
	Status string
	Sort   string
	Limit  int
	Offset int
	AdminView bool
}

func (r *OrderRepository) List(ctx context.Context, p ListOrdersParams) ([]models.Order, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if !p.AdminView {
		where = append(where, "user_id = ?")
		args = append(args, p.UserID)
	}
	if p.Status != "" {
		where = append(where, "status = ?")
		args = append(args, p.Status)
	}
	order := "created_at DESC"
	switch p.Sort {
	case "created_asc":
		order = "created_at ASC"
	case "created_desc":
		order = "created_at DESC"
	}
	query := `
		SELECT id, user_id, items, status, total_amount, created_at, updated_at
		FROM orders
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
	var res []models.Order
	for rows.Next() {
		var o models.Order
		var itemsStr, status, createdAt, updatedAt string
		if err := rows.Scan(&o.ID, &o.UserID, &itemsStr, &status, &o.TotalAmount, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal([]byte(itemsStr), &o.Items)
		o.Status = models.OrderStatus(status)
		o.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		o.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		res = append(res, o)
	}
	return res, rows.Err()
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id string, to models.OrderStatus) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx, `
		UPDATE orders SET status = ?, updated_at = ? WHERE id = ?
	`, string(to), now, id)
	return err
}

func (r *OrderRepository) Cancel(ctx context.Context, id string) error {
	return r.UpdateStatus(ctx, id, models.OrderStatusCancelled)
}

func CalculateTotal(items []models.OrderItem) (float64, error) {
	var total float64
	for _, it := range items {
		if it.Quantity <= 0 || it.Price < 0 {
			return 0, errors.New("invalid item")
		}
		total += float64(it.Quantity) * it.Price
	}
	return total, nil
}

func NewOrder(userID string, items []models.OrderItem) (models.Order, error) {
	total, err := CalculateTotal(items)
	if err != nil {
		return models.Order{}, err
	}
	return models.Order{
		ID:          uuid.NewString(),
		UserID:      userID,
		Items:       items,
		Status:      models.OrderStatusCreated,
		TotalAmount: total,
	}, nil
}


