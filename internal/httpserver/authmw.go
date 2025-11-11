package httpserver

import (
	"context"
	"net/http"
	"strings"

	"frame_control_system/internal/auth"
)

type authCtxKey struct{}

type AuthContext struct {
	UserID string
	Roles  []string
}

func AuthMiddleware(secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(h, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "missing bearer token"}})
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			claims, err := auth.ParseToken(token, secret)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, envelope{Success: false, Error: &apiError{Code: "unauthorized", Message: "invalid token"}})
				return
			}
			ctx := context.WithValue(r.Context(), authCtxKey{}, &AuthContext{
				UserID: claims.UserID,
				Roles:  claims.Roles,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ac := GetAuth(r)
			if ac == nil || !hasRole(ac.Roles, role) {
				writeJSON(w, http.StatusForbidden, envelope{Success: false, Error: &apiError{Code: "forbidden", Message: "insufficient permissions"}})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetAuth(r *http.Request) *AuthContext {
	if v := r.Context().Value(authCtxKey{}); v != nil {
		if ac, ok := v.(*AuthContext); ok {
			return ac
		}
	}
	return nil
}

func hasRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}


