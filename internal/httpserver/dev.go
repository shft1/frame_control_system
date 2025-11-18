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

type seedAdminRequest struct {
	Email    string `json:"email"`    // default admin@example.com
	Password string `json:"password"` // default admin123
	Name     string `json:"name"`     // default Admin
}

func DevSeedAdminHandler(db *sql.DB, jwtSecret string) http.HandlerFunc {
	userRepo := storage.NewUserRepository(db)
	return func(w http.ResponseWriter, r *http.Request) {
		var req seedAdminRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req.Email == "" {
			req.Email = "admin@example.com"
		}
		if req.Password == "" {
			req.Password = "admin123"
		}
		if req.Name == "" {
			req.Name = "Admin"
		}
		if _, err := mail.ParseAddress(req.Email); err != nil {
			writeJSON(w, http.StatusBadRequest, envelope{Success: false, Error: &apiError{Code: "invalid_input", Message: "invalid email"}})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		u, err := userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			// create new
			hash, err := auth.HashPassword(req.Password)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "hashing error"}})
				return
			}
			newUser := models.User{
				ID:           uuid.NewString(),
				Email:        strings.TrimSpace(req.Email),
				PasswordHash: hash,
				Name:         strings.TrimSpace(req.Name),
				Roles:        []string{"admin"},
			}
			if err := userRepo.Create(ctx, newUser); err != nil {
				writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "create error"}})
				return
			}
			token, _ := auth.GenerateToken(newUser.ID, newUser.Roles, jwtSecret, 24*time.Hour)
			writeJSON(w, http.StatusCreated, envelope{Success: true, Data: map[string]any{
				"user": map[string]any{"id": newUser.ID, "email": newUser.Email, "name": newUser.Name, "roles": newUser.Roles},
				"token": token,
			}})
			return
		}
		// ensure admin in roles
		rolesSet := map[string]struct{}{}
		for _, r := range u.Roles {
			rolesSet[r] = struct{}{}
		}
		rolesSet["admin"] = struct{}{}
		roles := make([]string, 0, len(rolesSet))
		for k := range rolesSet {
			roles = append(roles, k)
		}
		if err := userRepo.UpdateRoles(ctx, u.ID, roles); err != nil {
			writeJSON(w, http.StatusInternalServerError, envelope{Success: false, Error: &apiError{Code: "internal_error", Message: "update roles error"}})
			return
		}
		if req.Password != "" {
			hash, _ := auth.HashPassword(req.Password)
			_ = userRepo.UpdatePassword(ctx, u.ID, hash)
		}
		token, _ := auth.GenerateToken(u.ID, roles, jwtSecret, 24*time.Hour)
		writeJSON(w, http.StatusOK, envelope{Success: true, Data: map[string]any{
			"user": map[string]any{"id": u.ID, "email": u.Email, "name": u.Name, "roles": roles},
			"token": token,
		}})
	}
}


