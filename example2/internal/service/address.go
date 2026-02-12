package service

import (
	"context"
	"errors"
	"net/http"
	"reflect"

	gmodel "github.com/pilinux/gorest/database/model"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/repo"
)

// AddressService provides methods for address-related operations.
type AddressService struct {
	addressRepo repo.AddressRepository
}

// NewAddressService returns a new AddressService instance.
func NewAddressService(addressRepo repo.AddressRepository) *AddressService {
	return &AddressService{
		addressRepo: addressRepo,
	}
}

// AddAddress adds a new address to the database.
func (s *AddressService) AddAddress(ctx context.Context, address *model.Geocoding) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	res, err := s.addressRepo.AddAddress(ctx, address)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("AddAddress.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = res
	httpStatusCode = http.StatusCreated
	return
}

// GetAddresses retrieves all addresses from the database.
func (s *AddressService) GetAddresses(ctx context.Context) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	addr, err := s.addressRepo.GetAddresses(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("GetAddresses.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(addr) == 0 {
		httpResponse.Message = "no address found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = addr
	httpStatusCode = http.StatusOK
	return
}

// GetAddress retrieves an address by its ID.
func (s *AddressService) GetAddress(ctx context.Context, id string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	_id, err := bson.ObjectIDFromHex(id)
	if err != nil {
		resp := gmodel.HTTPResponse{
			Message: "invalid address ID format",
		}
		httpResponse = resp
		httpStatusCode = http.StatusBadRequest
		return
	}

	addr, err := s.addressRepo.GetAddress(ctx, _id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "address not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("GetAddress.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = addr
	httpStatusCode = http.StatusOK
	return
}

// GetAddressByFilter retrieves an address based on a filter.
func (s *AddressService) GetAddressByFilter(ctx context.Context, address *model.Geocoding, addDocIDInFilter bool) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	addr, err := s.addressRepo.GetAddressByFilter(ctx, address, addDocIDInFilter)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "address not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("GetAddressByFilter.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = addr
	httpStatusCode = http.StatusOK
	return
}

// UpdateAddress updates an existing address in the database.
func (s *AddressService) UpdateAddress(ctx context.Context, address *model.Geocoding) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if address == nil || address.ID.IsZero() {
		httpResponse.Message = "address ID is required"
		httpStatusCode = http.StatusBadRequest
		return
	}

	// check if the address exists
	existingAddress, err := s.addressRepo.GetAddress(ctx, address.ID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "address not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("UpdateAddress.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// if no changes are made, return the existing address
	if reflect.DeepEqual(existingAddress, address) {
		httpResponse.Message = existingAddress
		httpStatusCode = http.StatusOK
		return
	}

	err = s.addressRepo.UpdateAddressFields(ctx, address)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "address not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("UpdateAddress.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = address
	httpStatusCode = http.StatusOK
	return
}

// DeleteAddress deletes an address by its ID.
func (s *AddressService) DeleteAddress(ctx context.Context, id string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	_id, err := bson.ObjectIDFromHex(id)
	if err != nil {
		resp := gmodel.HTTPResponse{
			Message: "invalid address ID format",
		}
		httpResponse = resp
		httpStatusCode = http.StatusBadRequest
		return
	}

	err = s.addressRepo.DeleteAddress(ctx, _id)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "address not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("DeleteAddress.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = "address deleted successfully"
	httpStatusCode = http.StatusOK
	return
}
