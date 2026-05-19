package router

import (
	"net/http"

	authv1 "auth-service/gen/auth/v1"

	"github.com/gin-gonic/gin"
)

func getMe(authClient authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := userIDFromContext(c)
		if !ok {
			writeProblem(c, http.StatusUnauthorized, "Unauthorized", "authentication required")
			return
		}

		resp, err := authClient.GetUser(c.Request.Context(), &authv1.GetUserRequest{Id: userID})
		if err != nil {
			writeProblem(c, http.StatusInternalServerError, "Internal Server Error", "")
			return
		}

		c.JSON(http.StatusOK, resp.User)
	}
}
