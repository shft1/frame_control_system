package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type OutboxEvent struct {
	ID        string
	Type      string
	Payload   any
	CreatedAt time.Time
}

func AddOutboxEvent(ctx context.Context, db *sql.DB, eventType string, payload any) error {
	id := uuid.NewString()
	now := time.Now().UTC().Format(time.RFC3339)
	body, _ := json.Marshal(payload)
	_, err := db.ExecContext(ctx, `
		INSERT INTO outbox_events (id, type, payload, created_at)
		VALUES (?, ?, ?, ?)
	`, id, eventType, string(body), now)
	return err
}


