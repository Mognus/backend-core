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

type schema struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Fields      []field  `json:"fields"`
	Searchable  []string `json:"searchable"`
}

type field struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Label        string `json:"label"`
	Required     bool   `json:"required,omitempty"`
	Readonly     bool   `json:"readonly,omitempty"`
	TableHidden  bool   `json:"tableHidden,omitempty"`
	EditHidden   bool   `json:"editHidden,omitempty"`
	CreateHidden bool   `json:"createHidden,omitempty"`
}

func RegisterAdminRoutes(mux *http.ServeMux, mw func(http.Handler) http.Handler, service *Service) {
	registerResource(mux, mw, resource[Experience]{
		path:   "about-experiences",
		schema: experienceSchema(),
		list:   service.ListExperiences,
		get:    service.GetExperience,
		create: service.CreateExperience,
		save:   service.SaveExperience,
		delete: service.DeleteExperience,
	})
	registerResource(mux, mw, resource[Education]{
		path:   "about-education",
		schema: educationSchema(),
		list:   service.ListEducation,
		get:    service.GetEducation,
		create: service.CreateEducation,
		save:   service.SaveEducation,
		delete: service.DeleteEducation,
	})
	registerResource(mux, mw, resource[Skill]{
		path:   "about-skills",
		schema: skillSchema(),
		list:   service.ListSkills,
		get:    service.GetSkill,
		create: service.CreateSkill,
		save:   service.SaveSkill,
		delete: service.DeleteSkill,
	})
	registerResource(mux, mw, resource[Interest]{
		path:   "about-interests",
		schema: interestSchema(),
		list:   service.ListInterests,
		get:    service.GetInterest,
		create: service.CreateInterest,
		save:   service.SaveInterest,
		delete: service.DeleteInterest,
	})
}

type resource[T any] struct {
	path   string
	schema schema
	list   func(context.Context, grpccrud.ListRequest) ([]T, int64, error)
	get    func(context.Context, uint64) (T, error)
	create func(context.Context, *T) (*T, error)
	save   func(context.Context, uint64, *T) (*T, error)
	delete func(context.Context, uint64) error
}

func registerResource[T any](mux *http.ServeMux, mw func(http.Handler) http.Handler, res resource[T]) {
	base := "/api/admin/" + res.path

	mux.Handle("GET "+base+"/schema", mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, res.schema)
	})))
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

func experienceSchema() schema {
	return schema{
		Name:        "about-experiences",
		DisplayName: "About Experiences",
		Searchable:  []string{"company", "location", "active"},
		Fields: []field{
			{Name: "id", Type: "number", Label: "ID", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "company", Type: "string", Label: "Company", Required: true},
			{Name: "location", Type: "string", Label: "Location"},
			{Name: "startDate", Type: "date", Label: "Start Date", Required: true},
			{Name: "endDate", Type: "date", Label: "End Date"},
			{Name: "isCurrent", Type: "boolean", Label: "Current"},
			{Name: "sortOrder", Type: "number", Label: "Sort Order"},
			{Name: "active", Type: "boolean", Label: "Active"},
			{Name: "technologies", Type: "object", Label: "Technologies", TableHidden: true},
			{Name: "translations", Type: "object", Label: "Translations", TableHidden: true},
			{Name: "createdAt", Type: "date", Label: "Created", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true, EditHidden: true, CreateHidden: true},
		},
	}
}

func educationSchema() schema {
	return schema{
		Name:        "about-education",
		DisplayName: "About Education",
		Searchable:  []string{"institution", "location", "active"},
		Fields: []field{
			{Name: "id", Type: "number", Label: "ID", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "institution", Type: "string", Label: "Institution", Required: true},
			{Name: "location", Type: "string", Label: "Location"},
			{Name: "startDate", Type: "date", Label: "Start Date", Required: true},
			{Name: "endDate", Type: "date", Label: "End Date"},
			{Name: "isCurrent", Type: "boolean", Label: "Current"},
			{Name: "sortOrder", Type: "number", Label: "Sort Order"},
			{Name: "active", Type: "boolean", Label: "Active"},
			{Name: "translations", Type: "object", Label: "Translations", TableHidden: true},
			{Name: "createdAt", Type: "date", Label: "Created", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true, EditHidden: true, CreateHidden: true},
		},
	}
}

func skillSchema() schema {
	return schema{
		Name:        "about-skills",
		DisplayName: "About Skills",
		Searchable:  []string{"key", "category", "level", "active"},
		Fields: []field{
			{Name: "id", Type: "number", Label: "ID", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "key", Type: "string", Label: "Key", Required: true},
			{Name: "category", Type: "string", Label: "Category", Required: true},
			{Name: "level", Type: "string", Label: "Level", Required: true},
			{Name: "sortOrder", Type: "number", Label: "Sort Order"},
			{Name: "active", Type: "boolean", Label: "Active"},
			{Name: "translations", Type: "object", Label: "Translations", TableHidden: true},
			{Name: "createdAt", Type: "date", Label: "Created", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true, EditHidden: true, CreateHidden: true},
		},
	}
}

func interestSchema() schema {
	return schema{
		Name:        "about-interests",
		DisplayName: "About Interests",
		Searchable:  []string{"key", "active"},
		Fields: []field{
			{Name: "id", Type: "number", Label: "ID", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "key", Type: "string", Label: "Key", Required: true},
			{Name: "sortOrder", Type: "number", Label: "Sort Order"},
			{Name: "active", Type: "boolean", Label: "Active"},
			{Name: "translations", Type: "object", Label: "Translations", TableHidden: true},
			{Name: "createdAt", Type: "date", Label: "Created", Readonly: true, EditHidden: true, CreateHidden: true},
			{Name: "updatedAt", Type: "date", Label: "Updated", Readonly: true, EditHidden: true, CreateHidden: true},
		},
	}
}
