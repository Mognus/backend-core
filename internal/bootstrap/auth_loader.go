package bootstrap

import (
	"fmt"

	authclient "auth-service/client"
	"template/internal/admin"
)

type AuthLoader struct{}

func NewAuthLoader() AuthLoader {
	return AuthLoader{}
}

func (l AuthLoader) Load(runtime Runtime) error {
	authService, err := authclient.New(
		runtime.Config().Auth.ServiceAddr,
		runtime.Config().Auth.JWTSecret,
		runtime.Storage(),
	)
	if err != nil {
		return fmt.Errorf("connect auth-service: %w", err)
	}

	adminRegistry := admin.New(authService.Config, runtime.Router())
	runtime.SetProviderRegistrar(adminRegistry)
	runtime.ProviderRegistrar().RegisterProviders(authService)
	runtime.Services().RegisterServices(authService)

	return nil
}
