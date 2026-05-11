package about

import (
	"net/http"

	"template/internal/apierror"

	grpccrud "github.com/Mognus/go-grpc-crud/server"
)

func RegisterPublicRoutes(mux *http.ServeMux, service *Service) {
	mux.HandleFunc("GET /api/about/experiences", func(w http.ResponseWriter, r *http.Request) {
		items, _, err := service.ListExperiences(r.Context(), publicListRequest())
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		writeJSON(w, http.StatusOK, items)
	})
	mux.HandleFunc("GET /api/about/education", func(w http.ResponseWriter, r *http.Request) {
		items, _, err := service.ListEducation(r.Context(), publicListRequest())
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		writeJSON(w, http.StatusOK, items)
	})
	mux.HandleFunc("GET /api/about/skills", func(w http.ResponseWriter, r *http.Request) {
		items, _, err := service.ListSkills(r.Context(), publicListRequest())
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		writeJSON(w, http.StatusOK, items)
	})
	mux.HandleFunc("GET /api/about/interests", func(w http.ResponseWriter, r *http.Request) {
		items, _, err := service.ListInterests(r.Context(), publicListRequest())
		if err != nil {
			apierror.Internal(w, r)
			return
		}
		writeJSON(w, http.StatusOK, items)
	})
}

func publicListRequest() grpccrud.ListRequest {
	return grpccrud.ListRequest{
		Page:    1,
		Limit:   100,
		Filters: map[string]string{"active": "true"},
	}
}
