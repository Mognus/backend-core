package adminconf

import (
	"sort"

	"github.com/gofiber/fiber/v2"

	libcrud "github.com/Mognus/go-grpc-crud/crud"
)

type Module struct {
	providers       map[string]libcrud.GRPCProvider
	jwtMiddleware   fiber.Handler
	adminMiddleware fiber.Handler
}

type modelInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

func New() *Module {
	return &Module{
		providers: make(map[string]libcrud.GRPCProvider),
	}
}

func (m *Module) SetJWTMiddleware(middleware fiber.Handler) {
	m.jwtMiddleware = middleware
}

func (m *Module) SetAdminMiddleware(middleware fiber.Handler) {
	m.adminMiddleware = middleware
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

	if m.jwtMiddleware != nil {
		admin.Use(m.jwtMiddleware)
	}
	if m.adminMiddleware != nil {
		admin.Use(m.adminMiddleware)
	}

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
	return provider.SchemaHandler()(c)
}

func (m *Module) List(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.ListHandler()(c)
}

func (m *Module) Get(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.GetHandler()(c)
}

func (m *Module) Create(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.CreateHandler()(c)
}

func (m *Module) Update(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.UpdateHandler()(c)
}

func (m *Module) Delete(c *fiber.Ctx) error {
	provider, exists := m.getProvider(c)
	if !exists {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Model not found"})
	}
	return provider.DeleteHandler()(c)
}
