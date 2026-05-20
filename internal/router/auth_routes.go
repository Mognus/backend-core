package router

import (
	"context"
	"net/http"

	authv1 "auth-service/gen/auth/v1"

	"github.com/Mognus/go-grpc-crud/dbcrud"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type loginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type refreshInput struct {
	RefreshToken string `json:"refreshToken"`
}

type logoutInput struct {
	RefreshToken string `json:"refreshToken"`
}

type adminUserInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	RoleID    uint64 `json:"roleId"`
	Active    bool   `json:"active"`
}

type adminRoleInput struct {
	Name string `json:"name"`
}

type authRoutes struct {
	client authv1.AuthServiceClient
}

type adminUserDTO struct {
	ID        uint64        `json:"id"`
	Email     string        `json:"email"`
	FirstName string        `json:"firstName"`
	LastName  string        `json:"lastName"`
	RoleID    uint64        `json:"roleId"`
	Role      *adminRoleDTO `json:"role,omitempty"`
	Active    bool          `json:"active"`
	CreatedAt string        `json:"createdAt"`
	UpdatedAt string        `json:"updatedAt"`
}

type adminRoleDTO struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type authResponse struct {
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken,omitempty"`
	User         adminUserDTO `json:"user"`
}

type refreshResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

func newAuthRoutes(client authv1.AuthServiceClient) authRoutes {
	return authRoutes{client: client}
}

func (r authRoutes) RegisterRoutes(groups RouteGroups) {
	groups.Auth.POST("/login", login(r.client))
	groups.Auth.POST("/register", register(r.client))
	groups.Auth.POST("/refresh", refreshToken(r.client))
	groups.Auth.POST("/logout", logout(r.client))
	groups.AuthProtected.GET("/me", getMe(r.client))

	registerUserAdminRoutes(groups.Admin, r.client)
	registerRoleAdminRoutes(groups.Admin, r.client)
}

func login(authClient authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input loginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid JSON body")
			return
		}

		var header metadata.MD
		resp, err := authClient.Login(c.Request.Context(), &authv1.LoginRequest{
			Email:    input.Email,
			Password: input.Password,
		}, grpc.Header(&header))
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}

		forwardSetCookies(c, header)
		c.JSON(http.StatusOK, authResponse{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			User:         userDTO(resp.User),
		})
	}
}

func register(authClient authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input registerInput
		if err := c.ShouldBindJSON(&input); err != nil {
			writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid JSON body")
			return
		}

		var header metadata.MD
		resp, err := authClient.Register(c.Request.Context(), &authv1.RegisterRequest{
			Email:     input.Email,
			Password:  input.Password,
			FirstName: input.FirstName,
			LastName:  input.LastName,
		}, grpc.Header(&header))
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}

		forwardSetCookies(c, header)
		c.JSON(http.StatusCreated, authResponse{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			User:         userDTO(resp.User),
		})
	}
}

func refreshToken(authClient authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input refreshInput
		if c.Request.Body != nil {
			_ = c.ShouldBindJSON(&input)
		}

		var header metadata.MD
		ctx := contextWithCookies(c)
		resp, err := authClient.RefreshToken(ctx, &authv1.RefreshTokenRequest{
			RefreshToken: input.RefreshToken,
		}, grpc.Header(&header))
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}

		forwardSetCookies(c, header)
		c.JSON(http.StatusOK, refreshResponse{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
		})
	}
}

func logout(authClient authv1.AuthServiceClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input logoutInput
		if c.Request.Body != nil {
			_ = c.ShouldBindJSON(&input)
		}

		var header metadata.MD
		ctx := contextWithCookies(c)
		resp, err := authClient.Logout(ctx, &authv1.LogoutRequest{
			RefreshToken: input.RefreshToken,
		}, grpc.Header(&header))
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}

		forwardSetCookies(c, header)
		c.JSON(http.StatusOK, gin.H{"success": resp.Success})
	}
}

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

		c.JSON(http.StatusOK, userDTO(resp.User))
	}
}

func registerUserAdminRoutes(group *gin.RouterGroup, client authv1.AuthServiceClient) {
	group.GET("/users", func(c *gin.Context) {
		req := listUsersRequest(parseListRequest(c))
		resp, err := client.ListUsers(c.Request.Context(), req)
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": userDTOs(resp.Items), "total": resp.Total})
	})

	group.GET("/users/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		resp, err := client.GetUser(c.Request.Context(), &authv1.GetUserRequest{Id: id})
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": userDTO(resp.User)})
	})

	group.POST("/users", func(c *gin.Context) {
		var input adminUserInput
		if err := c.ShouldBindJSON(&input); err != nil {
			writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid JSON body")
			return
		}
		resp, err := client.CreateUser(c.Request.Context(), &authv1.CreateUserRequest{
			Email:     input.Email,
			Password:  input.Password,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			RoleId:    input.RoleID,
			Active:    input.Active,
		})
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": userDTO(resp.User)})
	})

	group.PUT("/users/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		var input adminUserInput
		if err := c.ShouldBindJSON(&input); err != nil {
			writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid JSON body")
			return
		}
		resp, err := client.UpdateUser(c.Request.Context(), &authv1.UpdateUserRequest{
			Id:        id,
			Email:     input.Email,
			Password:  input.Password,
			FirstName: input.FirstName,
			LastName:  input.LastName,
			RoleId:    input.RoleID,
			Active:    input.Active,
		})
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": userDTO(resp.User)})
	})

	group.DELETE("/users/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		if _, err := client.DeleteUser(c.Request.Context(), &authv1.DeleteUserRequest{Id: id}); err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id}})
	})
}

func registerRoleAdminRoutes(group *gin.RouterGroup, client authv1.AuthServiceClient) {
	group.GET("/roles", func(c *gin.Context) {
		req := listRolesRequest(parseListRequest(c))
		resp, err := client.ListRoles(c.Request.Context(), req)
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": roleDTOs(resp.Items), "total": resp.Total})
	})

	group.GET("/roles/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		resp, err := client.GetRole(c.Request.Context(), &authv1.GetRoleRequest{Id: id})
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": roleDTO(resp.Role)})
	})

	group.POST("/roles", func(c *gin.Context) {
		var input adminRoleInput
		if err := c.ShouldBindJSON(&input); err != nil {
			writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid JSON body")
			return
		}
		resp, err := client.CreateRole(c.Request.Context(), &authv1.CreateRoleRequest{Name: input.Name})
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusCreated, gin.H{"data": roleDTO(resp.Role)})
	})

	group.PUT("/roles/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		var input adminRoleInput
		if err := c.ShouldBindJSON(&input); err != nil {
			writeProblem(c, http.StatusBadRequest, "Bad Request", "invalid JSON body")
			return
		}
		resp, err := client.UpdateRole(c.Request.Context(), &authv1.UpdateRoleRequest{
			Id:   id,
			Name: input.Name,
		})
		if err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": roleDTO(resp.Role)})
	})

	group.DELETE("/roles/:id", func(c *gin.Context) {
		id, ok := parseID(c)
		if !ok {
			return
		}
		if _, err := client.DeleteRole(c.Request.Context(), &authv1.DeleteRoleRequest{Id: id}); err != nil {
			writeGRPCProblem(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": gin.H{"id": id}})
	})
}

func listUsersRequest(req dbcrud.ListRequest) *authv1.ListUsersRequest {
	return &authv1.ListUsersRequest{
		Page:      req.Page,
		Limit:     req.Limit,
		Search:    req.Search,
		Filters:   req.Filters,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}
}

func listRolesRequest(req dbcrud.ListRequest) *authv1.ListRolesRequest {
	return &authv1.ListRolesRequest{
		Page:      req.Page,
		Limit:     req.Limit,
		Search:    req.Search,
		Filters:   req.Filters,
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}
}

func userDTOs(users []*authv1.UserResponse) []adminUserDTO {
	items := make([]adminUserDTO, len(users))
	for i, user := range users {
		items[i] = userDTO(user)
	}
	return items
}

func userDTO(user *authv1.UserResponse) adminUserDTO {
	if user == nil {
		return adminUserDTO{}
	}
	return adminUserDTO{
		ID:        user.Id,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		RoleID:    user.RoleId,
		Role:      roleDTOPtr(user.Role),
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func roleDTOs(roles []*authv1.RoleResponse) []adminRoleDTO {
	items := make([]adminRoleDTO, len(roles))
	for i, role := range roles {
		items[i] = roleDTO(role)
	}
	return items
}

func roleDTO(role *authv1.RoleResponse) adminRoleDTO {
	if role == nil {
		return adminRoleDTO{}
	}
	return adminRoleDTO{
		ID:        role.Id,
		Name:      role.Name,
		CreatedAt: role.CreatedAt,
		UpdatedAt: role.UpdatedAt,
	}
}

func roleDTOPtr(role *authv1.RoleResponse) *adminRoleDTO {
	if role == nil || role.Id == 0 {
		return nil
	}
	dto := roleDTO(role)
	return &dto
}

func contextWithCookies(c *gin.Context) context.Context {
	ctx := c.Request.Context()
	if cookie := c.GetHeader("Cookie"); cookie != "" {
		return metadata.AppendToOutgoingContext(ctx, "cookie", cookie)
	}
	return ctx
}

func forwardSetCookies(c *gin.Context, header metadata.MD) {
	for _, cookie := range header.Get("set-cookie") {
		c.Writer.Header().Add("Set-Cookie", cookie)
	}
}
