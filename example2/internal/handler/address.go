package handler

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gmodel "github.com/pilinux/gorest/database/model"
	grenderer "github.com/pilinux/gorest/lib/renderer"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/service"
)

// AddressAPI provides HTTP handlers for address-related endpoints.
type AddressAPI struct {
	addressService *service.AddressService
}

// NewAddressAPI returns a new AddressAPI instance.
func NewAddressAPI(addressService *service.AddressService) *AddressAPI {
	return &AddressAPI{
		addressService: addressService,
	}
}

// AddAddress handles the HTTP POST request to add a new address.
//
// Endpoint: POST /api/v1/addresses
//
// Authorization: None
func (api *AddressAPI) AddAddress(c *gin.Context) {
	address := &model.Geocoding{}
	if err := c.ShouldBindJSON(address); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	address.Trim()
	if address.IsEmpty() {
		resp := gmodel.HTTPResponse{
			Message: "address is required",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.addressService.AddAddress(ctx, address)
	grenderer.Render(c, resp, statusCode)
}

// GetAddresses handles the HTTP GET request to retrieve all addresses.
//
// Endpoint: GET /api/v1/addresses
//
// Authorization: None
func (api *AddressAPI) GetAddresses(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.addressService.GetAddresses(ctx)
	grenderer.Render(c, resp, statusCode)
}

// GetAddress handles the HTTP GET request to retrieve an address by ID.
//
// Endpoint: GET /api/v1/addresses/:id
//
// Authorization: None
func (api *AddressAPI) GetAddress(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		resp := gmodel.HTTPResponse{
			Message: "address ID is required",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.addressService.GetAddress(ctx, id)
	grenderer.Render(c, resp, statusCode)
}

// GetAddressByFilter handles the HTTP POST request to retrieve an address by filter.
//
// Endpoint: POST /api/v1/addresses/filter?exclude-address-id=
//
// Authorization: None
func (api *AddressAPI) GetAddressByFilter(c *gin.Context) {
	address := &model.Geocoding{}
	if err := c.ShouldBindJSON(address); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	address.Trim()
	if address.IsEmpty() {
		resp := gmodel.HTTPResponse{
			Message: "address filter is empty",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	addDocIDInFilter := strings.ToLower(strings.TrimSpace(c.Query("exclude-address-id"))) != "true"
	resp, statusCode := api.addressService.GetAddressByFilter(ctx, address, addDocIDInFilter)
	grenderer.Render(c, resp, statusCode)
}

// UpdateAddress handles the HTTP PUT request to update an existing address.
//
// Endpoint: PUT /api/v1/addresses
//
// Authorization: None
func (api *AddressAPI) UpdateAddress(c *gin.Context) {
	address := &model.Geocoding{}
	if err := c.ShouldBindJSON(address); err != nil {
		grenderer.Render(c, gin.H{"message": err.Error()}, http.StatusBadRequest)
		return
	}
	address.Trim()
	if address.IsEmpty() {
		resp := gmodel.HTTPResponse{
			Message: "address is required",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.addressService.UpdateAddress(ctx, address)
	grenderer.Render(c, resp, statusCode)
}

// DeleteAddress handles the HTTP DELETE request to delete an address by ID.
//
// Endpoint: DELETE /api/v1/addresses/:id
//
// Authorization: None
func (api *AddressAPI) DeleteAddress(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		resp := gmodel.HTTPResponse{
			Message: "address ID is required",
		}
		grenderer.Render(c, resp, http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	resp, statusCode := api.addressService.DeleteAddress(ctx, id)
	grenderer.Render(c, resp, statusCode)
}
