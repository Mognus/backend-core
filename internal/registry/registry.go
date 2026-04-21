package registry

import "github.com/gofiber/fiber/v2"

type Service interface {
	Name() string
	Close()
}

type RoutableService interface {
	RegisterRoutes(fiber.Router)
}

type ServiceRegistry struct {
	router   fiber.Router
	services []Service
}

func New(router fiber.Router) *ServiceRegistry {
	return &ServiceRegistry{router: router}
}

func (r *ServiceRegistry) Add(services ...Service) {
	for _, service := range services {
		if routableService, ok := service.(RoutableService); ok {
			routableService.RegisterRoutes(r.router)
		}
	}

	r.services = append(r.services, services...)
}

func (r *ServiceRegistry) Close() {
	for _, s := range r.services {
		s.Close()
	}
}

func (r *ServiceRegistry) Names() []string {
	names := make([]string, len(r.services))
	for i, s := range r.services {
		names[i] = s.Name()
	}
	return names
}
