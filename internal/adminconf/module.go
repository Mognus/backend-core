package adminconf

import (
	"sort"

	"github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
)

type Module struct {
	auth      AdminAuth
	api       fiber.Router
	providers map[string]crud.GRPCProvider
}

type modelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type ProviderService interface {
	Providers() []crud.GRPCProvider
}

type AdminAuth interface {
	JWTMiddleware() fiber.Handler
	RequireAdmin(*fiber.Ctx) error
}

func New(auth AdminAuth, router fiber.Router) *Module {
	module := &Module{
		auth:      auth,
		providers: make(map[string]crud.GRPCProvider),
	}
	admin := router.Group("/admin")
	admin.Use(module.auth.JWTMiddleware())
	admin.Use(module.auth.RequireAdmin)

	module.api = admin.Group("/api")

	// Shared admin endpoints stay here; CRUD routes are mounted during provider registration.
	module.api.Get("/models", module.GetModels)

	return module
}

func (m *Module) RegisterCRUD(provider crud.GRPCProvider) {
	modelName := provider.GetModelName()
	if _, exists := m.providers[modelName]; exists {
		return
	}

	m.providers[modelName] = provider
	// Each provider mounts its own concrete admin routes under /admin/api/<model>.
	m.mountProviderRoutes(modelName, provider)
}

func (m *Module) RegisterProviders(services ...ProviderService) {
	for _, service := range services {
		for _, provider := range service.Providers() {
			m.RegisterCRUD(provider)
		}
	}
}

func (m *Module) mountProviderRoutes(modelName string, provider crud.GRPCProvider) {
	basePath := "/" + modelName

	m.api.Get(basePath, provider.HandleList)
	m.api.Post(basePath, provider.HandleCreate)
	m.api.Get(basePath+"/schema", provider.HandleSchema)
	m.api.Get(basePath+"/:id", provider.HandleGet)
	m.api.Put(basePath+"/:id", provider.HandleUpdate)
	m.api.Delete(basePath+"/:id", provider.HandleDelete)
}

func (m *Module) GetModels(c *fiber.Ctx) error {
	models := make([]modelInfo, 0, len(m.providers))

	for _, provider := range m.providers {
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
