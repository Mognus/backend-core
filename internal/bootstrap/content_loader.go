package bootstrap

import (
	"fmt"

	contentclient "content-service/client"
)

type ContentLoader struct{}

func NewContentLoader() ContentLoader {
	return ContentLoader{}
}

func (l ContentLoader) Load(runtime Runtime) error {
	contentService, err := contentclient.New(runtime.Config().Content.ServiceAddr)
	if err != nil {
		return fmt.Errorf("connect content-service: %w", err)
	}

	if runtime.ProviderRegistrar() == nil {
		contentService.Close()
		return fmt.Errorf("register content-service providers: provider registrar not initialized")
	}

	runtime.ProviderRegistrar().RegisterProviders(contentService)
	runtime.Services().RegisterServices(contentService)

	return nil
}
