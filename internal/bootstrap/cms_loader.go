package bootstrap

import (
	"fmt"

	cmsclient "cms-service/client"
)

type CMSLoader struct{}

func NewCMSLoader() CMSLoader {
	return CMSLoader{}
}

func (l CMSLoader) Load(runtime Runtime) error {
	cmsService, err := cmsclient.New(runtime.Config().CMS.ServiceAddr)
	if err != nil {
		return fmt.Errorf("connect cms-service: %w", err)
	}

	runtime.Services().RegisterServices(cmsService)

	return nil
}
