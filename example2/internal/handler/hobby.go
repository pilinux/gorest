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

// HobbyAPI provides HTTP handlers for hobby-related endpoints.
type HobbyAPI struct {
	hobbyService *service.HobbyService
}

// NewHobbyAPI returns a new HobbyAPI instance.
func NewHobbyAPI(hobbyService *service.HobbyService) *HobbyAPI {
	return &HobbyAPI{
		hobbyService: hobbyService,
	}
}

// GetHobbies handles the HTTP GET request to retrieve all hobbies.
//
// Endpoint: GET /api/v1/hobbies
//
// Authorization: None
func (api *HobbyAPI) GetHobbies(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.hobbyService.GetHobbies(ctx)
	grenderer.Render(c, resp, statusCode)
}

// GetHobby handles the HTTP GET request to retrieve a hobby by hobbyID.
//
// Endpoint: GET /api/v1/hobbies/:id
//
// Authorization: None
func (api *HobbyAPI) GetHobby(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		resp := gmodel.HTTPResponse{
			Message: "hobbyID is required",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	hobbyID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		resp := gmodel.HTTPResponse{
			Message: "invalid hobbyID format",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.hobbyService.GetHobby(ctx, hobbyID)
	grenderer.Render(c, resp, statusCode)
}

// GetHobbiesMe handles the HTTP GET request to retrieve hobbies for the authenticated user.
//
// Endpoint: GET /api/v1/hobbies/me
//
// Authorization: JWT token required
func (api *HobbyAPI) GetHobbiesMe(c *gin.Context) {
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

	resp, statusCode := api.hobbyService.GetHobbiesByAuthID(ctx, userIDAuth)
	grenderer.Render(c, resp, statusCode)
}

// AddHobbyToUser handles HTTP POST request to add a hobby to a user.
//
// Endpoint: POST /api/v1/hobbies
//
// Authorization: JWT token required
func (api *HobbyAPI) AddHobbyToUser(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		resp := gmodel.HTTPResponse{
			Message: "unauthorized",
		}
		grenderer.Render(c, resp, http.StatusUnauthorized)
		return
	}

	var hobby model.Hobby
	if err := c.ShouldBindJSON(&hobby); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	if err := hobby.Trim(); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.hobbyService.AddHobbyToUser(ctx, userIDAuth, &hobby)
	grenderer.Render(c, resp, statusCode)
}

// DeleteHobbyFromUser handles HTTP DELETE request to remove a hobby from a user.
//
// Endpoint: DELETE /api/v1/hobbies/:id
//
// Authorization: JWT token required
func (api *HobbyAPI) DeleteHobbyFromUser(c *gin.Context) {
	userIDAuth := getAuthID(c)
	if userIDAuth == 0 {
		resp := gmodel.HTTPResponse{
			Message: "unauthorized",
		}
		grenderer.Render(c, resp, http.StatusUnauthorized)
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		resp := gmodel.HTTPResponse{
			Message: "hobbyID is required",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	hobbyID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		resp := gmodel.HTTPResponse{
			Message: "invalid hobbyID format",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.hobbyService.DeleteHobbyFromUser(ctx, userIDAuth, hobbyID)
	grenderer.Render(c, resp, statusCode)
}
