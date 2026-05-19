package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func writeProblem(c *gin.Context, status int, title, detail string) {
	c.AbortWithStatusJSON(status, problem{
		Type:     "/problems/" + http.StatusText(status),
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: c.Request.URL.Path,
	})
}
