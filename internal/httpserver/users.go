package httpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"

	"frame_control_system/internal/auth"
	"frame_control_system/internal/models"
	"frame_control_system/internal/storage"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	userRepo := storage.NewUserRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "invalid json"}})
			return
		}
		req.Email = strings.TrimSpace(req.Email)
		req.Name = strings.TrimSpace(req.Name)
		if _, err := mail.ParseAddress(req.Email); err != nil || len(req.Password) < 6 || req.Name == "" {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "invalid email, password or name"}})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		exists, err := userRepo.EmailExists(ctx, req.Email)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "db error"}})
			return
		}
		if exists {
			writeJSON(w, http.StatusConflict, envelope{Success: false, Error: &apiError{Code: "email_taken", Message: "email already registered"}})
			return
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "hashing error"}})
			return
		}
		user := models.User{
			ID:           uuid.NewString(),
			Email:        req.Email,
			PasswordHash: hash,
			Name:         req.Name,
			Roles:        []string{"user"},
		}
		if err := userRepo.Create(ctx, user); err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "failed to create user"}})
			return
		}
		// Return minimal info
		writeJSON(w, http.StatusCreated, envelope{Success: true, Data: map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"roles": user.Roles,
		}})
	}
}

func LoginHandler(db *sql.DB, jwtSecret string) http.HandlerFunc {
	userRepo := storage.NewUserRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "invalid json"}})
			return
		}
		req.Email = strings.TrimSpace(req.Email)
		if req.Email == "" || req.Password == "" {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "email and password required"}})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		u, err := userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "invalid credentials"}})
			return
		}
		if !auth.CheckPassword(u.PasswordHash, req.Password) {
			writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "invalid credentials"}})
			return
		}
		token, err := auth.GenerateToken(u.ID, u.Roles, jwtSecret, 24*time.Hour)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "token error"}})
			return
		}
		writeJSON(w, http.StatusOK, envelope{Success: true, Data: map[string]interface{}{
			"token": token,
			"user": map[string]interface{}{
				"id":    u.ID,
				"email": u.Email,
				"name":  u.Name,
				"roles": u.Roles,
			},
		}})
	}
}


