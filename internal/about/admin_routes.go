package about

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"template/internal/apierror"

	grpccrud "github.com/Mognus/go-grpc-crud/server"
)

func RegisterAdminRoutes(mux *http.ServeMux, mw func(http.Handler) http.Handler, service *Service) {
	registerResource(mux, mw, resource[Experience]{
		path:   "about-experiences",
		list:   service.ListExperiences,
		get:    service.GetExperience,
		create: service.CreateExperience,
		save:   service.SaveExperience,
		delete: service.DeleteExperience,
	})
	registerResource(mux, mw, resource[Education]{
		path:   "about-education",
		list:   service.ListEducation,
		get:    service.GetEducation,
		create: service.CreateEducation,
		save:   service.SaveEducation,
		delete: service.DeleteEducation,
	})
	registerResource(mux, mw, resource[Skill]{
		path:   "about-skills",
		list:   service.ListSkills,
		get:    service.GetSkill,
		create: service.CreateSkill,
		save:   service.SaveSkill,
		delete: service.DeleteSkill,
	})
	registerResource(mux, mw, resource[Interest]{
		path:   "about-interests",
		list:   service.ListInterests,
		get:    service.GetInterest,
		create: service.CreateInterest,
		save:   service.SaveInterest,
		delete: service.DeleteInterest,
	})
}

type resource[T any] struct {
	path   string
	list   func(context.Context, grpccrud.ListRequest) ([]T, int64, error)
	get    func(context.Context, uint64) (T, error)
	create func(context.Context, *T) (*T, error)
	save   func(context.Context, uint64, *T) (*T, error)
	delete func(context.Context, uint64) error
}

func registerResource[T any](mux *http.ServeMux, mw func(http.Handler) http.Handler, res resource[T]) {
	base := "/api/admin/" + res.path

	mux.Handle("GET "+base, mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := parseListRequest(r)
		items, total, err := res.list(r.Context(), req)
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"items": items,
			"total": total,
			"page":  req.Page,
			"limit": req.Limit,
		})
	})))
	mux.Handle("GET "+base+"/{id}", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}
		item, err := res.get(r.Context(), id)
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})))
	mux.Handle("POST "+base, mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		item := new(T)
		if !decodeBody(w, r, item) {
			return
		}
		created, err := res.create(r.Context(), item)
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		writeJSON(w, http.StatusCreated, created)
	})))
	mux.Handle("PUT "+base+"/{id}", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}
		item := new(T)
		if !decodeBody(w, r, item) {
			return
		}
		updated, err := res.save(r.Context(), id, item)
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		writeJSON(w, http.StatusOK, updated)
	})))
	mux.Handle("DELETE "+base+"/{id}", mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := parseID(w, r)
		if !ok {
			return
		}
		if err := res.delete(r.Context(), id); err != nil {
			apierror.Internal(w, r)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})))
}

func parseListRequest(r *http.Request) grpccrud.ListRequest {
	query := r.URL.Query()
	filters := make(map[string]string)
	for key, values := range query {
		if !strings.HasPrefix(key, "filters[") || !strings.HasSuffix(key, "]") || len(values) == 0 {
			continue
		}
		filters[strings.TrimSuffix(strings.TrimPrefix(key, "filters["), "]")] = values[0]
	}
	return grpccrud.ListRequest{
		Page:    int32(parseInt(query.Get("page"), 1)),
		Limit:   int32(parseInt(query.Get("limit"), 20)),
		Filters: filters,
	}
}

func parseID(w http.ResponseWriter, r *http.Request) (uint64, bool) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		apierror.BadRequest(w, r, "invalid id")
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

func decodeBody[T any](w http.ResponseWriter, r *http.Request, target *T) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		apierror.BadRequest(w, r, "invalid JSON body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	body, _ := json.Marshal(value)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(body)
}
