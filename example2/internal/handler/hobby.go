package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
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
	resp, statusCode := api.hobbyService.GetHobbies()
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
		grenderer.Render(c, "hobbyID is required", http.StatusBadRequest)
		return
	}

	hobbyID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		grenderer.Render(c, "invalid hobbyID format", http.StatusBadRequest)
		return
	}

	resp, statusCode := api.hobbyService.GetHobby(hobbyID)
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
		grenderer.Render(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, statusCode := api.hobbyService.GetHobbiesByAuthID(userIDAuth)
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
		grenderer.Render(c, "unauthorized", http.StatusUnauthorized)
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

	resp, statusCode := api.hobbyService.AddHobbyToUser(userIDAuth, &hobby)
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
		grenderer.Render(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		grenderer.Render(c, "hobbyID is required", http.StatusBadRequest)
		return
	}

	hobbyID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		grenderer.Render(c, "invalid hobbyID format", http.StatusBadRequest)
		return
	}

	resp, statusCode := api.hobbyService.DeleteHobbyFromUser(userIDAuth, hobbyID)
	grenderer.Render(c, resp, statusCode)
}
