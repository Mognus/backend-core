package main

import (
	"log"

	authclient "auth-service/client"
	"template/internal/admin"
	"template/internal/services"

	"github.com/gofiber/fiber/v2"
	fiberredis "github.com/gofiber/storage/redis/v2"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment")
	}

	redisStorage := fiberredis.New(fiberredis.Config{
		URL: getEnv("REDIS_URL", "redis://localhost:6379"),
	})

	app := newApp()
	api := app.Group("/api")
	serviceRegistry := services.New(api)

	loadServices(api, serviceRegistry, redisStorage)
	defer serviceRegistry.Close()

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"modules": serviceRegistry.Names(),
		})
	})

	addr := getEnv("HOST", "0.0.0.0") + ":" + getEnv("PORT", "8080")
	log.Printf("Server starting on http://%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func loadServices(router fiber.Router, serviceRegistry *services.ServiceRegistry, storage fiber.Storage) {
	authSvc, err := authclient.New(
		getEnv("AUTH_SERVICE_ADDR", "localhost:50051"),
		getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		storage,
	)
	if err != nil {
		log.Fatalf("Failed to connect to auth-service: %v", err)
	}

	adminRegistry := admin.New(authSvc.Config, router)
	adminRegistry.RegisterProviders(authSvc)
	serviceRegistry.RegisterServices(authSvc)
}
