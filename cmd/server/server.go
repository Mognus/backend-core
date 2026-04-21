package main

import (
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"template/internal/config"

	apperrors "github.com/Mognus/go-grpc-crud/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func newApp(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
		BodyLimit:    50 * 1024 * 1024,
	})
	app.Use(recover.New(recover.Config{EnableStackTrace: true}))
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORS.AllowOrigins,
		AllowCredentials: true,
	}))
	app.Static("/uploads", "./uploads")
	return app
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
