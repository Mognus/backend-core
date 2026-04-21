package admin

import (
	"github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
)

type ProviderService interface {
	Providers() []crud.GRPCProvider
}

type AdminAuth interface {
	JWTMiddleware() fiber.Handler
	RequireAdmin(*fiber.Ctx) error
}
