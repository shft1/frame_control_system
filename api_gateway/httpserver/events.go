package httpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type outboxEventDTO struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Payload   map[string]any         `json:"payload"`
	CreatedAt string                 `json:"created_at"`
}

func AdminListOutboxHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		limit := parseIntDefault(q.Get("limit"), 50, 1, 500)
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		rows, err := db.QueryContext(ctx, `
			SELECT id, type, payload, created_at
			FROM outbox_events
			ORDER BY created_at DESC
			LIMIT ?
		`, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "db error"}})
			return
		}
		defer rows.Close()
		var res []outboxEventDTO
		for rows.Next() {
			var e outboxEventDTO
			var payloadStr string
			if err := rows.Scan(&e.ID, &e.Type, &payloadStr, &e.CreatedAt); err != nil {
				writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "scan error"}})
				return
			}
			_ = json.Unmarshal([]byte(payloadStr), &e.Payload)
			res = append(res, e)
		}
		writeJSON(w, http.StatusOK, envelope{Success: true, Data: map[string]any{
			"items": res,
			"limit": limit,
		}})
	}
}


