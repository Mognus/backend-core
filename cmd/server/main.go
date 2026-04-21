package main

import (
	"log"

	"template/internal/bootstrap"
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

	runtime := bootstrap.NewRuntime(cfg, api, redisStorage, serviceRegistry)
	if err := runtime.Load(bootstrap.NewAuthLoader()); err != nil {
		log.Fatalf("Failed to load services: %v", err)
	}
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
