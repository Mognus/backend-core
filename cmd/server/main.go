package main

import (
	"log"

	authclient "auth-service/client"
	"template/internal/adminconf"
	serviceregistry "template/internal/registry"

	libcrud "github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
	fiberredis "github.com/gofiber/storage/redis/v2"
	"github.com/joho/godotenv"
)

type FiberService interface {
	RegisterRoutes(fiber.Router)
}

type CRUDProviderService interface {
	Providers() []libcrud.GRPCProvider
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	redisStorage := fiberredis.New(fiberredis.Config{
		URL: getEnv("REDIS_URL", "redis://localhost:6379"),
	})

	app := newApp()
	api := app.Group("/api")
	services := serviceregistry.New()

	loadServices(api, services, redisStorage)
	defer services.Close()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"modules": services.Names(),
		})
	})

	addr := getEnv("HOST", "0.0.0.0") + ":" + getEnv("PORT", "8080")
	log.Printf("Server starting on http://%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func registerServiceRuntime(router fiber.Router, admin *adminconf.Module, services ...serviceregistry.Service) {
	for _, service := range services {
		if fiberService, ok := service.(FiberService); ok {
			fiberService.RegisterRoutes(router)
		}

		if providerService, ok := service.(CRUDProviderService); ok {
			for _, provider := range providerService.Providers() {
				admin.RegisterCRUD(provider)
			}
		}
	}
}

func loadServices(router fiber.Router, services *serviceregistry.ServiceRegistry, storage fiber.Storage) {
	authSvc, err := authclient.New(
		getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		storage,
	)
	if err != nil {
		log.Fatalf("Failed to connect to auth-service: %v", err)
	}

	admin := adminconf.New(authSvc.Config)
	admin.Mount(router)

	runtimeServices := []serviceregistry.Service{authSvc}

	registerServiceRuntime(router, admin, runtimeServices...)
	services.Add(runtimeServices...)
}
