package apierror

import (
	"encoding/json"
	"net/http"
)

type problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func write(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	p := problem{
		Type:     "/problems/" + http.StatusText(status),
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: r.URL.Path,
	}
	body, _ := json.Marshal(p)
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	w.Write(body)
}

func Unauthorized(w http.ResponseWriter, r *http.Request, detail string) {
	write(w, r, http.StatusUnauthorized, "Unauthorized", detail)
}

func Forbidden(w http.ResponseWriter, r *http.Request, detail string) {
	write(w, r, http.StatusForbidden, "Forbidden", detail)
}

func BadRequest(w http.ResponseWriter, r *http.Request, detail string) {
	write(w, r, http.StatusBadRequest, "Bad Request", detail)
}

func Internal(w http.ResponseWriter, r *http.Request) {
	write(w, r, http.StatusInternalServerError, "Internal Server Error", "")
}
