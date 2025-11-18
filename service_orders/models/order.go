package models

import "time"

type OrderItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

type OrderStatus string

const (
	OrderStatusCreated    OrderStatus = "created"
	OrderStatusInProgress OrderStatus = "in_progress"
	OrderStatusDone       OrderStatus = "done"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID          string       `json:"id"`
	UserID      string       `json:"user_id"`
	Items       []OrderItem  `json:"items"`
	Status      OrderStatus  `json:"status"`
	TotalAmount float64      `json:"total_amount"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}


