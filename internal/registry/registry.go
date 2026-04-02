package registry

import (
	authclient "auth-service/client"
	cmsclient "cms-service/client"
	galleryclient "gallery-service/client"
	"template/internal/adminconf"

	libcrud "github.com/Mognus/go-grpc-crud/crud"
	"github.com/gofiber/fiber/v2"
)

type client interface {
	Name() string
	RegisterRoutes(fiber.Router)
	Providers() []libcrud.GRPCProvider
	Close()
}

type Registry struct {
	auth    *authclient.AuthService
	clients []client
	admin   *adminconf.Module
}

func New() *Registry {
	return &Registry{admin: adminconf.New()}
}

func (r *Registry) register(c client) {
	r.clients = append(r.clients, c)
	for _, p := range c.Providers() {
		r.admin.RegisterCRUD(p)
	}
}

func (r *Registry) AddAuth(addr, jwtSecret string, storage fiber.Storage) error {
	svc, err := authclient.New(addr, jwtSecret, storage)
	if err != nil {
		return err
	}
	r.auth = svc
	r.admin.SetJWTMiddleware(svc.Config.JWTMiddleware())
	r.admin.SetAdminMiddleware(svc.Config.RequireAdmin)
	r.register(svc)
	return nil
}

func (r *Registry) AddGallery(addr string) error {
	svc, err := galleryclient.New(addr)
	if err != nil {
		return err
	}
	r.register(svc)
	return nil
}

func (r *Registry) AddCMS(addr string) error {
	svc, err := cmsclient.New(addr, r.auth.Config)
	if err != nil {
		return err
	}
	r.register(svc)
	return nil
}

func (r *Registry) Mount(router fiber.Router) {
	for _, c := range r.clients {
		c.RegisterRoutes(router)
	}
	r.admin.Mount(router)
}

func (r *Registry) CloseAll() {
	for _, c := range r.clients {
		c.Close()
	}
}

func (r *Registry) Names() []string {
	names := make([]string, len(r.clients))
	for i, c := range r.clients {
		names[i] = c.Name()
	}
	return names
}
