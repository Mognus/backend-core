package services

import "github.com/gofiber/fiber/v2"

type Service interface {
	Name() string
	Close()
}

type RoutableService interface {
	RegisterRoutes(fiber.Router)
}
