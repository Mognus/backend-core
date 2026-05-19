package router

import "github.com/gin-gonic/gin"

type routeModule interface {
	RegisterPublicRoutes(api *gin.RouterGroup)
	RegisterAdminRoutes(admin *gin.RouterGroup)
}

func registerModules(api *gin.RouterGroup, admin *gin.RouterGroup, modules ...routeModule) {
	for _, module := range modules {
		module.RegisterPublicRoutes(api)
		module.RegisterAdminRoutes(admin)
	}
}
