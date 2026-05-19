// Package gateway wires all gRPC services into a single HTTP handler via grpc-gateway.
//
// Flow: HTTP request → net/http ServeMux → grpc-gateway runtime.ServeMux
//
//	→ generated handler (from proto HTTP annotations) → gRPC client → service.
//
// Route precedence (Go 1.22 ServeMux): more specific patterns win.
// /me is registered before the public catch-all so it takes priority.
package gateway

import (
	"context"
	"net/http"
	"strings"

	authv1 "auth-service/gen/auth/v1"

	"template/internal/config"
	"template/internal/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Gateway struct {
	Handler    http.Handler
	AuthClient authv1.AuthServiceClient
	cleanup    func()
}

func (g *Gateway) Close() {
	if g.cleanup != nil {
		g.cleanup()
	}
}

// New builds the gateway handler and owns the gRPC clients it creates.
func New(ctx context.Context, cfg *config.Config) (*Gateway, error) {
	authConn, err := grpc.NewClient(cfg.Auth.ServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	cleanup := func() {
		authConn.Close()
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
		return nil, err
	}

	jwtMw := middleware.JWT(cfg.Auth.JWTSecret)
	// adminMw chains JWT validation + role check into a single middleware.
	adminMw := func(h http.Handler) http.Handler { return jwtMw(middleware.RequireAdmin(h)) }

	mux := http.NewServeMux()

	// Admin routes are protected; gwMux handles the actual routing internally.
	mux.Handle("/api/admin/", adminMw(gwMux))

	// Public catch-all — covers /api/auth/login, /api/auth/register, etc.
	mux.Handle("/", gwMux)

	return &Gateway{
		Handler:    mux,
		AuthClient: authClient,
		cleanup:    cleanup,
	}, nil
}
