package httpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"frame_control_system/service_orders/events"
	orderModels "frame_control_system/service_orders/models"
	orderStorage "frame_control_system/service_orders/storage"
)

type createOrderRequest struct {
	Items []orderModels.OrderItem `json:"items"`
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func CreateOrderHandler(db *sql.DB) http.HandlerFunc {
	repo := orderStorage.NewOrderRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		ac := GetAuth(r)
		if ac == nil {
			writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "no auth"}})
			return
		}
		var req createOrderRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.Items) == 0 {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "invalid items"}})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		order, err := orderStorage.NewOrder(ac.UserID, req.Items)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "invalid order items"}})
			return
		}
		if err := repo.Create(ctx, order); err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "db error"}})
			return
		}
		_ = orderStorage.AddOutboxEvent(ctx, db, events.OrderCreated, map[string]any{
			"id":      order.ID,
			"user_id": order.UserID,
			"status":  order.Status,
			"total":   order.TotalAmount,
		})
		writeJSON(w, http.StatusCreated, envelope{Success: true, Data: order})
	}
}

func GetOrderHandler(db *sql.DB) http.HandlerFunc {
	repo := orderStorage.NewOrderRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		ac := GetAuth(r)
		if ac == nil {
			writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "no auth"}})
			return
		}
		// Prefer chi URLParam
		id := chi.URLParam(r, "id")
		if id == "" {
			id = strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
		}
		if id == "" {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "id required"}})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		o, err := repo.GetByID(ctx, id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, envelope{Success: false, Error: &apiError{Code: "not_found", Message: "order not found"}})
			return
		}
		if o.UserID != ac.UserID && !hasRole(ac.Roles, "admin") {
			writeJSON(w, http.StatusForbidden, envelope{Success: false, Error: &apiError{Code: "forbidden", Message: "not allowed"}})
			return
		}
		writeJSON(w, http.StatusOK, envelope{Success: true, Data: o})
	}
}

func ListOrdersHandler(db *sql.DB) http.HandlerFunc {
	repo := orderStorage.NewOrderRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		ac := GetAuth(r)
		if ac == nil {
			writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "no auth"}})
			return
		}
		q := r.URL.Query()
		limit := parseIntDefault(q.Get("limit"), 20, 1, 100)
		page := parseIntDefault(q.Get("page"), 1, 1, 100000)
		offset := (page - 1) * limit
		status := strings.TrimSpace(q.Get("status"))
		sort := strings.TrimSpace(q.Get("sort"))
		params := orderStorage.ListOrdersParams{
			UserID:    ac.UserID,
			Status:    status,
			Sort:      sort,
			Limit:     limit,
			Offset:    offset,
			AdminView: hasRole(ac.Roles, "admin"),
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		list, err := repo.List(ctx, params)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "db error"}})
			return
		}
		writeJSON(w, http.StatusOK, envelope{Success: true, Data: map[string]any{
			"items": list,
			"page":  page,
			"limit": limit,
		}})
	}
}

func UpdateOrderStatusHandler(db *sql.DB) http.HandlerFunc {
	repo := orderStorage.NewOrderRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		ac := GetAuth(r)
		if ac == nil {
			writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "no auth"}})
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			tmp := strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
			id = strings.TrimSuffix(tmp, "/status")
		}
		if id == "" {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "id required"}})
			return
		}
		var req updateStatusRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "invalid json"}})
			return
		}
		to := orderModels.OrderStatus(strings.TrimSpace(req.Status))
		if to != orderModels.OrderStatusInProgress && to != orderModels.OrderStatusDone {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "unsupported status"}})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		o, err := repo.GetByID(ctx, id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, envelope{Success: false, Error: &apiError{Code: "not_found", Message: "order not found"}})
			return
		}
		if o.UserID != ac.UserID && !hasRole(ac.Roles, "admin") {
			writeJSON(w, http.StatusForbidden, envelope{Success: false, Error: &apiError{Code: "forbidden", Message: "not allowed"}})
			return
		}
		if err := validateTransition(o.Status, to); err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_transition", Message: err.Error()}})
			return
		}
		if err := repo.UpdateStatus(ctx, id, to); err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "db error"}})
			return
		}
		_ = orderStorage.AddOutboxEvent(ctx, db, events.OrderStatusUpdate, map[string]any{
			"id":     o.ID,
			"status": to,
		})
		o.Status = to
		writeJSON(w, http.StatusOK, envelope{Success: true, Data: o})
	}
}

func CancelOrderHandler(db *sql.DB) http.HandlerFunc {
	repo := orderStorage.NewOrderRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		ac := GetAuth(r)
		if ac == nil {
			writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "no auth"}})
			return
		}
		id := chi.URLParam(r, "id")
		if id == "" {
			id = strings.TrimPrefix(r.URL.Path, "/api/v1/orders/")
		}
		if id == "" {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "id required"}})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		o, err := repo.GetByID(ctx, id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, envelope{Success: false, Error: &apiError{Code: "not_found", Message: "order not found"}})
			return
		}
		if o.UserID != ac.UserID && !hasRole(ac.Roles, "admin") {
			writeJSON(w, http.StatusForbidden, envelope{Success: false, Error: &apiError{Code: "forbidden", Message: "not allowed"}})
			return
		}
		// allowed cancel from created or in_progress
		if err := validateTransition(o.Status, orderModels.OrderStatusCancelled); err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_transition", Message: err.Error()}})
			return
		}
		if err := repo.Cancel(ctx, id); err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "db error"}})
			return
		}
		_ = orderStorage.AddOutboxEvent(ctx, db, events.OrderStatusUpdate, map[string]any{
			"id":     o.ID,
			"status": orderModels.OrderStatusCancelled,
		})
		o.Status = orderModels.OrderStatusCancelled
		writeJSON(w, http.StatusOK, envelope{Success: true, Data: o})
	}
}

func validateTransition(from, to orderModels.OrderStatus) error {
	switch from {
	case orderModels.OrderStatusCreated:
		if to == orderModels.OrderStatusInProgress || to == orderModels.OrderStatusCancelled {
			return nil
		}
	case orderModels.OrderStatusInProgress:
		if to == orderModels.OrderStatusDone || to == orderModels.OrderStatusCancelled {
			return nil
		}
	case orderModels.OrderStatusDone, orderModels.OrderStatusCancelled:
		// terminal
	}
	return errors.New(fmt.Sprintf("cannot transition from %s to %s", from, to))
}


