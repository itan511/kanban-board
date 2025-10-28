package middleware

import (
	"context"
	"net/http"
	"strings"

	"kanban-board/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	CtxUserEmail ctxKey = "userEmail"
)

func NewAuthMiddleware(jwtKey []byte) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				http.Error(w, "authorization header required", http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(auth, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			tokenString := parts[1]

			var claims types.Claims
			token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrTokenUnverifiable
				}
				return jwtKey, nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			if strings.TrimSpace(claims.Email) == "" {
				http.Error(w, "token missing email claim", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxUserEmail, claims.Email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(CtxUserEmail).(string)
	return email, ok
}
