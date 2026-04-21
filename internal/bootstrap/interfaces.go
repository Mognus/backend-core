package bootstrap

import (
	"template/internal/admin"
	"template/internal/config"
	"template/internal/services"

	"github.com/gofiber/fiber/v2"
)

type ServiceRegistrar interface {
	RegisterServices(...services.Service)
}

type ProviderRegistrar interface {
	RegisterProviders(...admin.ProviderService)
}

type Runtime interface {
	Config() *config.Config
	Router() fiber.Router
	Storage() fiber.Storage
	Services() ServiceRegistrar
	Providers() ProviderRegistrar
}

type Loader interface {
	Load(Runtime) error
}
