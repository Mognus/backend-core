package registry

import (
	authclient "auth-service/client"
	"template/internal/adminconf"

	libcrud "github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
)

type Service interface {
	Name() string
	RegisterRoutes(fiber.Router)
	Providers() []libcrud.GRPCProvider
	Close()
}

type Registry struct {
	router   fiber.Router
	auth     *authclient.AuthService
	services []Service
	admin    *adminconf.Module
}

func New(router fiber.Router) *Registry {
	return &Registry{router: router}
}

func (r *Registry) SetAuth(s *authclient.AuthService) {
	r.auth = s
	r.admin = adminconf.New(s.Config)
	r.admin.Mount(r.router)
	r.registerProviders(s)
	r.registerRoutes(s)
	r.services = append(r.services, s)
}

func (r *Registry) AddService(s Service) {
	r.registerProviders(s)
	r.registerRoutes(s)
	r.services = append(r.services, s)
}

func (r *Registry) registerProviders(s Service) {
	for _, p := range s.Providers() {
		r.admin.RegisterCRUD(p)
	}
}

func (r *Registry) registerRoutes(s Service) {
	s.RegisterRoutes(r.router)
}

func (r *Registry) CloseAll() {
	for _, s := range r.services {
		s.Close()
	}
}

func (r *Registry) Names() []string {
	names := make([]string, len(r.services))
	for i, s := range r.services {
		names[i] = s.Name()
	}
	return names
}
