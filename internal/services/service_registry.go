package services

import "github.com/gofiber/fiber/v2"

type ServiceRegistry struct {
	router   fiber.Router
	services []Service
}

func New(router fiber.Router) *ServiceRegistry {
	return &ServiceRegistry{router: router}
}

func (r *ServiceRegistry) RegisterServices(services ...Service) {
	r.services = append(r.services, services...)

	for _, service := range services {
		if routableService, ok := service.(RoutableService); ok {
			// Routes are wired when a service is registered; the registry also tracks shutdown and names.
			routableService.RegisterRoutes(r.router)
		}
	}
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
