package main

import (
	"encoding/json"
	stderrors "errors"
	"log"
	"log/slog"
	"os"

	"template/internal/registry"

	apperrors "github.com/Mognus/go-grpc-crud/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
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

	r := registry.New()

	if err := r.AddAuth(getEnv("AUTH_SERVICE_ADDR", "localhost:50051"), getEnv("JWT_SECRET", "your-secret-key-change-in-production"), redisStorage); err != nil {
		log.Fatalf("Failed to connect to auth-service: %v", err)
	}
	if err := r.AddGallery(getEnv("GALLERY_SERVICE_ADDR", "localhost:50052")); err != nil {
		log.Fatalf("Failed to connect to gallery-service: %v", err)
	}
	if err := r.AddCMS(getEnv("CMS_SERVICE_ADDR", "localhost:50053")); err != nil {
		log.Fatalf("Failed to connect to cms-service: %v", err)
	}
	defer r.CloseAll()

	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
		BodyLimit:    50 * 1024 * 1024,
	})
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000"),
		AllowCredentials: true,
	}))
	app.Static("/uploads", "./uploads")

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"modules": r.Names(),
		})
	})

	api := app.Group("/api")
	r.Mount(api)

	addr := getEnv("HOST", "0.0.0.0") + ":" + getEnv("PORT", "8080")
	log.Printf("Server starting on http://%s", addr)
	if err := app.Listen(addr); err != nil {
		log.Fatal(err)
	}
}

func errorHandler(c *fiber.Ctx, err error) error {
	var problem *apperrors.Problem
	if stderrors.As(err, &problem) {
		problem.Instance = c.Path()
	} else {
		var fiberErr *fiber.Error
		if stderrors.As(err, &fiberErr) {
			problem = &apperrors.Problem{
				Type:     "/problems/http-error",
				Title:    fiberErr.Message,
				Status:   fiberErr.Code,
				Instance: c.Path(),
			}
		} else {
			problem = &apperrors.Problem{
				Type:     "/problems/internal-error",
				Title:    "Internal Server Error",
				Status:   500,
				Detail:   "An unexpected error occurred",
				Instance: c.Path(),
			}
		}
	}
	if problem.Status >= 500 {
		slog.Error(c.Method()+" "+c.Path(), "status", problem.Status, "error", err)
	}
	body, _ := json.Marshal(problem)
	c.Set(fiber.HeaderContentType, "application/problem+json")
	return c.Status(problem.Status).Send(body)
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
