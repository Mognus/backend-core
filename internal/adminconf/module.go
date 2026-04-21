package adminconf

import (
	"sort"

	authclient "auth-service/client"
	libcrud "github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
)

type Module struct {
	auth      *authclient.Config
	api       fiber.Router
	providers map[string]libcrud.GRPCProvider
	services  []ProviderService
}

type modelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type ProviderService interface {
	Providers() []libcrud.GRPCProvider
}

func New(auth *authclient.Config) *Module {
	return &Module{
		auth:      auth,
		providers: make(map[string]libcrud.GRPCProvider),
	}
}

func (m *Module) RegisterCRUD(provider libcrud.GRPCProvider) {
	modelName := provider.GetModelName()
	if _, exists := m.providers[modelName]; exists {
		return
	}

	m.providers[modelName] = provider
	// Each provider mounts its own concrete admin routes under /admin/api/<model>.
	m.mountProviderRoutes(modelName, provider)
}

func (m *Module) RegisterProviders(services ...ProviderService) {
	if m.api == nil {
		panic("admin module must be mounted before registering providers")
	}

	m.services = append(m.services, services...)

	for _, service := range services {
		for _, provider := range service.Providers() {
			m.RegisterCRUD(provider)
		}
	}
}

func (m *Module) Mount(router fiber.Router) {
	admin := router.Group("/admin")
	admin.Use(m.auth.JWTMiddleware())
	admin.Use(m.auth.RequireAdmin)

	m.api = admin.Group("/api")

	// Shared admin endpoints stay here; CRUD routes are mounted during provider registration.
	m.api.Get("/models", m.GetModels)
}

func (m *Module) mountProviderRoutes(modelName string, provider libcrud.GRPCProvider) {
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
