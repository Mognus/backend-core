package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func writeProblem(c *gin.Context, status int, title, detail string) {
	c.AbortWithStatusJSON(status, problem{
		Type:     "/problems/" + http.StatusText(status),
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: c.Request.URL.Path,
	})
}

func writeGRPCProblem(c *gin.Context, err error) {
	st, ok := status.FromError(err)
	if !ok {
		writeProblem(c, http.StatusInternalServerError, "Internal Server Error", "")
		return
	}

	httpStatus := httpStatusFromCode(st.Code())
	detail := ""
	if st.Code() != codes.Internal {
		detail = st.Message()
	}
	writeProblem(c, httpStatus, http.StatusText(httpStatus), detail)
}

func httpStatusFromCode(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument, codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
