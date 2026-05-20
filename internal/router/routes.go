package router

import "github.com/gin-gonic/gin"

type RouteGroups struct {
	API           *gin.RouterGroup
	Auth          *gin.RouterGroup
	AuthProtected *gin.RouterGroup
	Admin         *gin.RouterGroup
}

type Routes interface {
	RegisterRoutes(groups RouteGroups)
}

func registerRoutes(groups RouteGroups, routes ...Routes) {
	for _, route := range routes {
		route.RegisterRoutes(groups)
	}
}
