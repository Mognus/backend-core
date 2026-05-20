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

func (r aboutRoutes) RegisterRoutes(groups RouteGroups) {
	registerAboutPublicRoutes(groups.API.Group("/about"), r.service)
	registerAboutAdminRoutes(groups.Admin, r.service)
}

func registerAboutPublicRoutes(group *gin.RouterGroup, service *about.Service) {
	group.GET("/experiences", publicList(service.ListExperiences))
	group.GET("/education", publicList(service.ListEducation))
	group.GET("/skills", publicList(service.ListSkills))
	group.GET("/interests", publicList(service.ListInterests))
}

func registerAboutAdminRoutes(group *gin.RouterGroup, service *about.Service) {
	registerAdminResource(group, adminResource[about.Experience]{
		path:   "about-experiences",
		list:   service.ListExperiences,
		get:    service.GetExperience,
		create: service.CreateExperience,
		save:   service.SaveExperience,
		delete: service.DeleteExperience,
	})
	registerAdminResource(group, adminResource[about.Education]{
		path:   "about-education",
		list:   service.ListEducation,
		get:    service.GetEducation,
		create: service.CreateEducation,
		save:   service.SaveEducation,
		delete: service.DeleteEducation,
	})
	registerAdminResource(group, adminResource[about.Skill]{
		path:   "about-skills",
		list:   service.ListSkills,
		get:    service.GetSkill,
		create: service.CreateSkill,
		save:   service.SaveSkill,
		delete: service.DeleteSkill,
	})
	registerAdminResource(group, adminResource[about.Interest]{
		path:   "about-interests",
		list:   service.ListInterests,
		get:    service.GetInterest,
		create: service.CreateInterest,
		save:   service.SaveInterest,
		delete: service.DeleteInterest,
	})
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
