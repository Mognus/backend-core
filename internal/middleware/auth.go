package middleware

import (
	"context"
	"net/http"
	"strings"

	"template/internal/apierror"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsKey contextKey = "claims"

func JWT(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				apierror.Unauthorized(w, r, "missing token")
				return
			}

			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				apierror.Unauthorized(w, r, "invalid token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				apierror.Unauthorized(w, r, "invalid claims")
				return
			}

			ctx := context.WithValue(r.Context(), claimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := ClaimsFromContext(r.Context())
		if !ok {
			apierror.Unauthorized(w, r, "authentication required")
			return
		}
		if role, _ := claims["role"].(string); role != "admin" {
			apierror.Forbidden(w, r, "admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ClaimsFromContext(ctx context.Context) (jwt.MapClaims, bool) {
	claims, ok := ctx.Value(claimsKey).(jwt.MapClaims)
	return claims, ok
}

func UserIDFromContext(ctx context.Context) (uint64, bool) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return 0, false
	}
	id, ok := claims["user_id"].(float64)
	return uint64(id), ok
}

func extractToken(r *http.Request) string {
	if h := r.Header.Get("Authorization"); h != "" {
		// Accept both "Bearer <token>" (browser) and raw token (server-side Next.js fetch).
		return strings.TrimPrefix(h, "Bearer ")
	}
	if c, err := r.Cookie("access_token"); err == nil {
		return c.Value
	}
	return ""
}
