package service

import (
	"context"
	"errors"
	"net/http"

	gmodel "github.com/pilinux/gorest/database/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/repo"
)

// HobbyService provides methods for hobby-related operations.
type HobbyService struct {
	hobbyRepo repo.HobbyRepository
	userRepo  repo.UserRepository
}

// NewHobbyService returns a new HobbyService instance.
func NewHobbyService(hobbyRepo repo.HobbyRepository, userRepo repo.UserRepository) *HobbyService {
	return &HobbyService{
		hobbyRepo: hobbyRepo,
		userRepo:  userRepo,
	}
}

// GetHobbies retrieves all hobbies.
func (s *HobbyService) GetHobbies(ctx context.Context) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	hobbies, err := s.hobbyRepo.GetHobbies(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("GetHobbies.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(hobbies) == 0 {
		httpResponse.Message = "no hobby found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = hobbies
	httpStatusCode = http.StatusOK
	return
}

// GetHobby retrieves a hobby with the given hobbyID.
func (s *HobbyService) GetHobby(ctx context.Context, hobbyID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	hobby, err := s.hobbyRepo.GetHobby(ctx, hobbyID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "hobby not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("GetHobby.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = hobby
	httpStatusCode = http.StatusOK
	return
}

// GetHobbiesByAuthID retrieves hobbies for a user by their authID.
func (s *HobbyService) GetHobbiesByAuthID(ctx context.Context, authID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	user, err := s.userRepo.GetUserByAuthID(ctx, authID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "no user profile found"
			httpStatusCode = http.StatusForbidden
			return
		}

		log.WithError(err).Error("GetHobbiesByAuthID.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	hobbies, err := s.hobbyRepo.GetHobbiesByUserID(ctx, user.UserID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "no hobbies found for this user"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("GetHobbiesByAuthID.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = hobbies
	httpStatusCode = http.StatusOK
	return
}

// AddHobbyToUser adds a hobby to a user.
func (s *HobbyService) AddHobbyToUser(ctx context.Context, authID uint64, hobby *model.Hobby) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	user, err := s.userRepo.GetUserByAuthID(ctx, authID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "no user profile found"
			httpStatusCode = http.StatusForbidden
			return
		}

		log.WithError(err).Error("AddHobbyToUser.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if err := s.hobbyRepo.AddHobbyToUser(ctx, hobby, user); err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("AddHobbyToUser.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = hobby
	httpStatusCode = http.StatusOK
	return
}

// DeleteHobbyFromUser removes a hobby from a user.
func (s *HobbyService) DeleteHobbyFromUser(ctx context.Context, authID, hobbyID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	user, err := s.userRepo.GetUserByAuthID(ctx, authID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "no user profile found"
			httpStatusCode = http.StatusForbidden
			return
		}

		log.WithError(err).Error("DeleteHobbyFromUser.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if err := s.hobbyRepo.DeleteHobbyFromUser(ctx, hobbyID, user); err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "hobby not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("DeleteHobbyFromUser.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = "hobby deleted successfully"
	httpStatusCode = http.StatusOK
	return
}
