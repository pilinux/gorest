// Package handler provides HTTP handler implementations for the application's API endpoints.
// It defines handler types and functions that process incoming HTTP requests and interact with the service layer.
package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gmodel "github.com/pilinux/gorest/database/model"
	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/service"
)

// UserAPI provides HTTP handlers for user-related endpoints.
type UserAPI struct {
	userService *service.UserService
}

// NewUserAPI returns a new UserAPI instance.
func NewUserAPI(userService *service.UserService) *UserAPI {
	return &UserAPI{
		userService: userService,
	}
}

// GetUsers handles the HTTP GET request to retrieve all users and their posts.
//
// Endpoint: GET /api/v1/users
//
// Authorization: None
func (api *UserAPI) GetUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.userService.GetUsers(ctx)
	grenderer.Render(c, resp, statusCode)
}

// GetUser handles the HTTP GET request to retrieve a user by userID and their posts.
//
// Endpoint: GET /api/v1/users/:id
//
// Authorization: None
func (api *UserAPI) GetUser(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		resp := gmodel.HTTPResponse{
			Message: "userID is required",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		resp := gmodel.HTTPResponse{
			Message: "invalid userID format",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.userService.GetUser(ctx, userID)
	grenderer.Render(c, resp, statusCode)
}

// CreateUser handles the HTTP POST request to create a new user.
//
// Endpoint: POST /api/v1/users
//
// Authorization: JWT token required
func (api *UserAPI) CreateUser(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		resp := gmodel.HTTPResponse{
			Message: "unauthorized",
		}
		grenderer.Render(c, resp, http.StatusUnauthorized)
		return
	}

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := user.Trim(); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	user.IDAuth = userIDAuth

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.userService.CreateUser(ctx, &user)
	grenderer.Render(c, resp, statusCode)
}

// UpdateUser handles the HTTP PUT request to update an existing user.
//
// Endpoint: PUT /api/v1/users
//
// Authorization: JWT token required
func (api *UserAPI) UpdateUser(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		resp := gmodel.HTTPResponse{
			Message: "unauthorized",
		}
		grenderer.Render(c, resp, http.StatusUnauthorized)
		return
	}

	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := user.Trim(); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	user.IDAuth = userIDAuth

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.userService.UpdateUser(ctx, &user)
	grenderer.Render(c, resp, statusCode)
}

// DeleteUser handles the HTTP DELETE request to delete a user account permanently.
// This operation is irreversible.
//
// Endpoint: DELETE /api/v1/users
//
// Authorization: JWT token required
func (api *UserAPI) DeleteUser(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		resp := gmodel.HTTPResponse{
			Message: "unauthorized",
		}
		grenderer.Render(c, resp, http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.userService.DeleteUser(ctx, userIDAuth)
	grenderer.Render(c, resp, statusCode)
}
