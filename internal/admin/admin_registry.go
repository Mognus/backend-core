package admin

import (
	"sort"

	"github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
)

type AdminRegistry struct {
	auth      AdminAuth
	api       fiber.Router
	providers map[string]crud.GRPCProvider
}

type modelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func New(auth AdminAuth, router fiber.Router) *AdminRegistry {
	adminRegistry := &AdminRegistry{
		auth:      auth,
		providers: make(map[string]crud.GRPCProvider),
	}
	admin := router.Group("/admin")
	admin.Use(adminRegistry.auth.JWTMiddleware())
	admin.Use(adminRegistry.auth.RequireAdmin)

	adminRegistry.api = admin.Group("/api")

	// Shared admin endpoints stay here; CRUD routes are mounted during provider registration.
	adminRegistry.api.Get("/models", adminRegistry.GetModels)

	return adminRegistry
}

func (a *AdminRegistry) registerCRUD(provider crud.GRPCProvider) {
	modelName := provider.GetModelName()
	if _, exists := a.providers[modelName]; exists {
		return
	}

	a.providers[modelName] = provider
	// Each provider mounts its own concrete admin routes under /admin/api/<model>.
	a.mountProviderRoutes(modelName, provider)
}

func (a *AdminRegistry) RegisterProviders(services ...ProviderService) {
	for _, service := range services {
		for _, provider := range service.Providers() {
			a.registerCRUD(provider)
		}
	}
}

func (a *AdminRegistry) mountProviderRoutes(modelName string, provider crud.GRPCProvider) {
	basePath := "/" + modelName

	a.api.Get(basePath, provider.HandleList)
	a.api.Post(basePath, provider.HandleCreate)
	a.api.Get(basePath+"/schema", provider.HandleSchema)
	a.api.Get(basePath+"/:id", provider.HandleGet)
	a.api.Put(basePath+"/:id", provider.HandleUpdate)
	a.api.Delete(basePath+"/:id", provider.HandleDelete)
}

func (a *AdminRegistry) GetModels(c *fiber.Ctx) error {
	models := make([]modelInfo, 0, len(a.providers))

	for _, provider := range a.providers {
		schema := provider.GetSchema()
		models = append(models, modelInfo{
			Name:        schema.Name,
			DisplayName: schema.DisplayName,
		})
	}

	sort.Slice(models, func(i, j int) bool {
		if models[i].DisplayName == models[j].DisplayName {
			return models[i].Name < models[j].Name
		}
		return models[i].DisplayName < models[j].DisplayName
	})

	return c.JSON(fiber.Map{
		"models": models,
	})
}
