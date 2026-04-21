package main

import (
	"log"

	authclient "auth-service/client"
	"template/internal/admin"
	"template/internal/config"
	"template/internal/services"

	"github.com/gofiber/fiber/v2"
	fiberredis "github.com/gofiber/storage/redis/v2"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	redisStorage := fiberredis.New(fiberredis.Config{
		URL: cfg.Redis.URL,
	})

	app := newApp(cfg)
	api := app.Group("/api")
	serviceRegistry := services.New(api)

	loadServices(cfg, api, serviceRegistry, redisStorage)
	defer serviceRegistry.Close()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"modules": serviceRegistry.Names(),
		})
	})

	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Printf("Server starting on http://%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func loadServices(cfg *config.Config, router fiber.Router, serviceRegistry *services.ServiceRegistry, storage fiber.Storage) {
	authSvc, err := authclient.New(
		cfg.Auth.ServiceAddr,
		cfg.Auth.JWTSecret,
		storage,
	)
	if err != nil {
		log.Fatalf("Failed to connect to auth-service: %v", err)
	}

	adminRegistry := admin.New(authSvc.Config, router)
	adminRegistry.RegisterProviders(authSvc)
	serviceRegistry.RegisterServices(authSvc)
}
