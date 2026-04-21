package adminconf

import (
	"sort"

	authclient "auth-service/client"
	libcrud "github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
)

type Module struct {
	auth      *authclient.Config
	providers map[string]libcrud.GRPCProvider
}

type modelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func New(auth *authclient.Config) *Module {
	return &Module{
		auth:      auth,
		providers: make(map[string]libcrud.GRPCProvider),
	}
}

func (m *Module) RegisterCRUD(provider libcrud.GRPCProvider) {
	m.providers[provider.GetModelName()] = provider
}

func (m *Module) getProvider(c *fiber.Ctx) (libcrud.GRPCProvider, bool) {
	modelName := c.Params("model")
	provider, exists := m.providers[modelName]
	if !exists {
		return nil, false
	}
	return provider, true
}

func (m *Module) Mount(router fiber.Router) {
	admin := router.Group("/admin")
	admin.Use(m.auth.JWTMiddleware())
	admin.Use(m.auth.RequireAdmin)

	api := admin.Group("/api")

	api.Get("/models", m.GetModels)
	api.Get("/:model", m.List)
	api.Get("/:model/schema", m.GetSchema)
	api.Get("/:model/:id", m.Get)
	api.Post("/:model", m.Create)
	api.Put("/:model/:id", m.Update)
	api.Delete("/:model/:id", m.Delete)
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

func (m *Module) GetSchema(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.HandleSchema(c)
}

func (m *Module) List(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.HandleList(c)
}

func (m *Module) Get(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.HandleGet(c)
}

func (m *Module) Create(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.HandleCreate(c)
}

func (m *Module) Update(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.HandleUpdate(c)
}

func (m *Module) Delete(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.HandleDelete(c)
}
