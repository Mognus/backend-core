package router

import (
	"context"
	"net/http"

	"template/internal/about"

	"github.com/Mognus/go-grpc-crud/dbcrud"
	"github.com/gin-gonic/gin"
)

type aboutRoutes struct {
	service *about.Service
}

func newAboutRoutes(service *about.Service) aboutRoutes {
	return aboutRoutes{service: service}
}

func (r aboutRoutes) RegisterPublicRoutes(api *gin.RouterGroup) {
	registerAboutPublicRoutes(api.Group("/about"), r.service)
}

func (r aboutRoutes) RegisterAdminRoutes(admin *gin.RouterGroup) {
	registerAboutAdminRoutes(admin, r.service)
}

func registerAboutPublicRoutes(group *gin.RouterGroup, service *about.Service) {
	group.GET("/experiences", publicList(service.ListExperiences))
	group.GET("/education", publicList(service.ListEducation))
	group.GET("/skills", publicList(service.ListSkills))
	group.GET("/interests", publicList(service.ListInterests))
}

func publicList[T any](list func(context.Context, dbcrud.ListRequest) ([]T, int64, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		items, _, err := list(c.Request.Context(), publicListRequest())
		if err != nil {
			writeProblem(c, http.StatusInternalServerError, "Internal Server Error", "")
			return
		}
		c.JSON(http.StatusOK, items)
	}
}

func publicListRequest() dbcrud.ListRequest {
	return dbcrud.ListRequest{
		Page:    1,
		Limit:   100,
		Filters: map[string]string{"active": "true"},
	}
}
