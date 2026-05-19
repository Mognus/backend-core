package router

import (
	"net/http"
	"strings"

	authv1 "auth-service/gen/auth/v1"

	"template/internal/config"

	"github.com/gin-gonic/gin"
)

type Deps struct {
	Gateway    http.Handler
	AuthClient authv1.AuthServiceClient
}

func New(cfg *config.Config, deps Deps) http.Handler {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors(cfg.CORS.AllowOrigins))

	api := r.Group("/api")
	auth := api.Group("/auth")
	auth.GET("/me", jwtAuth(cfg.Auth.JWTSecret), getMe(deps.AuthClient))

	r.NoRoute(gin.WrapH(deps.Gateway))
	r.NoMethod(gin.WrapH(deps.Gateway))

	return r
}

func cors(allowOrigins string) gin.HandlerFunc {
	allowed := parseAllowedOrigins(allowOrigins)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func parseAllowedOrigins(value string) map[string]bool {
	allowed := make(map[string]bool)
	for _, origin := range strings.Split(value, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return allowed
}
