// Package gateway wires all gRPC services into a single HTTP handler via grpc-gateway.
//
// Flow: HTTP request → net/http ServeMux → grpc-gateway runtime.ServeMux
//
//	→ generated handler (from proto HTTP annotations) → gRPC client → service.
//
// Route precedence (Go 1.22 ServeMux): more specific patterns win.
// Schema and /me are registered first so they take priority over the admin catch-all.
package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	authv1 "auth-service/gen/auth/v1"
	cmsv1 "cms-service/gen/cms/v1"

	"template/internal/about"
	"template/internal/apierror"
	"template/internal/config"
	"template/internal/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"
)

// New builds the gateway handler and returns a cleanup func to close gRPC connections.
func New(ctx context.Context, cfg *config.Config, database *gorm.DB) (http.Handler, func(), error) {
	authConn, err := grpc.NewClient(cfg.Auth.ServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	cmsConn, err := grpc.NewClient(cfg.CMS.ServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		authConn.Close()
		return nil, nil, err
	}

	cleanup := func() {
		authConn.Close()
		cmsConn.Close()
	}

	authClient := authv1.NewAuthServiceClient(authConn)

	// gwMux translates HTTP → gRPC using routes defined in proto HTTP annotations.
	// Cookie forwarding is enabled so the auth service can read refresh_token from
	// incoming gRPC metadata on /refresh and /logout.
	gwMux := runtime.NewServeMux(
		runtime.WithErrorHandler(errorHandler),
		// Forward Cookie header inbound so gRPC handlers can read refresh_token from metadata.
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if strings.ToLower(key) == "cookie" {
				return key, true
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
		// Forward set-cookie outbound — grpc-gateway drops it by default.
		runtime.WithOutgoingHeaderMatcher(func(key string) (string, bool) {
			if strings.ToLower(key) == "set-cookie" {
				return "Set-Cookie", true
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
	)

	// RegisterXxxServiceHandlerClient wires all RPCs annotated with google.api.http
	// onto gwMux. One call per service — no manual route registration needed.
	if err := authv1.RegisterAuthServiceHandlerClient(ctx, gwMux, authClient); err != nil {
		cleanup()
		return nil, nil, err
	}
	if err := cmsv1.RegisterCmsServiceHandlerClient(ctx, gwMux, cmsv1.NewCmsServiceClient(cmsConn)); err != nil {
		cleanup()
		return nil, nil, err
	}

	jwtMw := middleware.JWT(cfg.Auth.JWTSecret)
	// adminMw chains JWT validation + role check into a single middleware.
	adminMw := func(h http.Handler) http.Handler { return jwtMw(middleware.RequireAdmin(h)) }

	mux := http.NewServeMux()

	// /api/admin/models replaces the old AdminRegistry endpoint. Model names carry
	// the service prefix (e.g. "auth/users") so the frontend builds URLs generically.
	type modelInfo struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
	}
	models := []modelInfo{
		{"users", "Users"},
		{"roles", "Roles"},
		{"content-groups", "Content Groups"},
		{"about-experiences", "About Experiences"},
		{"about-education", "About Education"},
		{"about-skills", "About Skills"},
		{"about-interests", "About Interests"},
	}
	modelsBody, _ := json.Marshal(map[string]any{"models": models})
	mux.HandleFunc("GET /api/admin/models", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(modelsBody)
	})

	// Core-owned admin models are registered directly; gRPC-backed admin models
	// continue to be served by gwMux below.
	aboutService := about.NewService(database)
	about.RegisterAdminRoutes(mux, adminMw, aboutService)

	// All schema endpoints are served via grpc-gateway RPCs in each service.

	// /me is the only custom handler: GetUser requires an ID which we extract from
	// the JWT — grpc-gateway can't do that without a dedicated proto RPC.
	mux.Handle("GET /api/auth/me", jwtMw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.UserIDFromContext(r.Context())
		if !ok {
			apierror.Unauthorized(w, r, "authentication required")
			return
		}
		resp, err := authClient.GetUser(r.Context(), &authv1.GetUserRequest{Id: userID})
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.User)
	})))

	// Admin routes are protected; gwMux handles the actual routing internally.
	mux.Handle("/api/admin/", adminMw(gwMux))

	// Public catch-all — covers /api/auth/login, /api/auth/register, /api/cms/content, etc.
	mux.Handle("/", gwMux)

	return corsMiddleware(cfg.CORS.AllowOrigins, mux), cleanup, nil
}

func corsMiddleware(allowOrigins string, next http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, o := range strings.Split(allowOrigins, ",") {
		allowed[strings.TrimSpace(o)] = true
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && allowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
