package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const claimsKey = "claims"

func jwtAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			writeProblem(c, http.StatusUnauthorized, "Unauthorized", "missing token")
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			writeProblem(c, http.StatusUnauthorized, "Unauthorized", "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeProblem(c, http.StatusUnauthorized, "Unauthorized", "invalid claims")
			return
		}

		c.Set(claimsKey, claims)
		c.Next()
	}
}

func requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := claimsFromContext(c)
		if !ok {
			writeProblem(c, http.StatusUnauthorized, "Unauthorized", "authentication required")
			return
		}
		if role, _ := claims["role"].(string); role != "admin" {
			writeProblem(c, http.StatusForbidden, "Forbidden", "admin access required")
			return
		}
		c.Next()
	}
}

func claimsFromContext(c *gin.Context) (jwt.MapClaims, bool) {
	value, ok := c.Get(claimsKey)
	if !ok {
		return nil, false
	}
	claims, ok := value.(jwt.MapClaims)
	return claims, ok
}

func userIDFromContext(c *gin.Context) (uint64, bool) {
	claims, ok := claimsFromContext(c)
	if !ok {
		return 0, false
	}
	id, ok := claims["user_id"].(float64)
	return uint64(id), ok
}

func extractToken(c *gin.Context) string {
	if h := c.GetHeader("Authorization"); h != "" {
		return strings.TrimPrefix(h, "Bearer ")
	}
	if cookie, err := c.Cookie("access_token"); err == nil {
		return cookie
	}
	return ""
}
