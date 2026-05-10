package gateway

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type problem struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

func errorHandler(_ context.Context, _ *runtime.ServeMux, _ runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	st, _ := status.FromError(err)
	httpStatus := runtime.HTTPStatusFromCode(st.Code())

	p := problem{
		Type:     problemType(st.Code()),
		Title:    http.StatusText(httpStatus),
		Status:   httpStatus,
		Instance: r.URL.Path,
	}
	if st.Code() != codes.Internal {
		p.Detail = st.Message()
	}

	body, _ := json.Marshal(p)
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(httpStatus)
	w.Write(body)
}

func problemType(code codes.Code) string {
	switch code {
	case codes.NotFound:
		return "/problems/not-found"
	case codes.Unauthenticated:
		return "/problems/unauthorized"
	case codes.PermissionDenied:
		return "/problems/forbidden"
	case codes.InvalidArgument, codes.FailedPrecondition:
		return "/problems/bad-request"
	case codes.AlreadyExists:
		return "/problems/conflict"
	case codes.ResourceExhausted:
		return "/problems/too-many-requests"
	default:
		return "/problems/internal-error"
	}
}
