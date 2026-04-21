package bootstrap

import (
	"template/internal/config"

	"github.com/gofiber/fiber/v2"
)

type AppRuntime struct {
	config    *config.Config
	router    fiber.Router
	storage   fiber.Storage
	services  ServiceRegistrar
	providers ProviderRegistrar
}

func NewRuntime(cfg *config.Config, router fiber.Router, storage fiber.Storage, services ServiceRegistrar) *AppRuntime {
	return &AppRuntime{
		config:   cfg,
		router:   router,
		storage:  storage,
		services: services,
	}
}

func (r *AppRuntime) Config() *config.Config {
	return r.config
}

func (r *AppRuntime) Router() fiber.Router {
	return r.router
}

func (r *AppRuntime) Storage() fiber.Storage {
	return r.storage
}

func (r *AppRuntime) Services() ServiceRegistrar {
	return r.services
}

func (r *AppRuntime) ProviderRegistrar() ProviderRegistrar {
	return r.providers
}

func (r *AppRuntime) SetProviderRegistrar(providers ProviderRegistrar) {
	r.providers = providers
}

func (r *AppRuntime) Load(loaders ...Loader) error {
	for _, loader := range loaders {
		if err := loader.Load(r); err != nil {
			return err
		}
	}

	return nil
}
