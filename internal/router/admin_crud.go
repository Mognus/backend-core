package router

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/Mognus/go-grpc-crud/dbcrud"
	"github.com/gin-gonic/gin"
)

type adminResource[T any] struct {
	path   string
	list   func(context.Context, dbcrud.ListRequest) ([]T, int64, error)
	get    func(context.Context, uint64) (T, error)
	create func(context.Context, *T) (*T, error)
	save   func(context.Context, uint64, *T) (*T, error)
	delete func(context.Context, uint64) error
}

func registerAdminResource[T any](group *gin.RouterGroup, res adminResource[T]) {
	group.GET("/"+res.path, func(c *gin.Context) {
		req := parseListRequest(c)
		items, total, err := res.list(c.Request.Context(), req)
		if err != nil {
			writeProblem(c, http.StatusInternalServerError, "Internal Server Error", "")
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": items, "total": total})
	})

	group.GET("/"+res.path+"/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		item, err := res.get(c.Request.Context(), id)
		if err != nil {
			writeProblem(c, http.StatusInternalServerError, "Internal Server Error", "")
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": item})
	})

	group.POST("/"+res.path, func(c *gin.Context) {
		item := new(T)
		if err := c.ShouldBindJSON(item); err != nil {
			writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid JSON body")
			return
		}
		created, err := res.create(c.Request.Context(), item)
		if err != nil {
			writeProblem(c, http.StatusInternalServerError, "Internal Server Error", "")
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": created})
	})

	group.PUT("/"+res.path+"/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		item := new(T)
		if err := c.ShouldBindJSON(item); err != nil {
			writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid JSON body")
			return
		}
		updated, err := res.save(c.Request.Context(), id, item)
		if err != nil {
			writeProblem(c, http.StatusInternalServerError, "Internal Server Error", "")
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": updated})
	})

	group.DELETE("/"+res.path+"/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		if err := res.delete(c.Request.Context(), id); err != nil {
			writeProblem(c, http.StatusInternalServerError, "Internal Server Error", "")
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id}})
	})
}

func parseListRequest(c *gin.Context) dbcrud.ListRequest {
	filters := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if !strings.HasPrefix(key, "filters[") || !strings.HasSuffix(key, "]") || len(values) == 0 {
			continue
		}
		filters[strings.TrimSuffix(strings.TrimPrefix(key, "filters["), "]")] = values[0]
	}

	pageSize := parseInt(c.Query("pageSize"), 0)
	if pageSize == 0 {
		pageSize = parseInt(c.Query("limit"), 20)
	}

	return dbcrud.ListRequest{
		Page:      int32(parseInt(c.Query("page"), 1)),
		Limit:     int32(pageSize),
		Search:    c.Query("search"),
		Filters:   filters,
		SortBy:    c.Query("sort"),
		SortOrder: c.Query("order"),
	}
}

func parseID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid id")
		return 0, false
	}
	return id, true
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
