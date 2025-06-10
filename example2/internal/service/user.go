// Package service provides business logic and application services for the API.
// It defines service types and methods that interact with repositories and implement core operations.
package service

import (
	"net/http"

	gdb "github.com/pilinux/gorest/database"
	gmodel "github.com/pilinux/gorest/database/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/pilinux/gorest/example2/internal/database/model"
	"github.com/pilinux/gorest/example2/internal/repo"
)

// UserService provides methods for user-related operations.
type UserService struct {
	userRepo  repo.UserRepository
	postRepo  repo.PostRepository
	hobbyRepo repo.HobbyRepository
}

// NewUserService returns a new UserService instance.
func NewUserService(userRepo repo.UserRepository, postRepo repo.PostRepository, hobbyRepo repo.HobbyRepository) *UserService {
	return &UserService{
		userRepo:  userRepo,
		postRepo:  postRepo,
		hobbyRepo: hobbyRepo,
	}
}

// GetUsers returns all users along with their posts.
func (s *UserService) GetUsers() (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	users, err := s.userRepo.GetUsers()
	if err != nil {
		log.WithError(err).Error("GetUsers.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(users) == 0 {
		httpResponse.Message = "no user found"
		httpStatusCode = http.StatusNotFound
		return
	}

	// fetch posts for each user
	for j, user := range users {
		posts, err := s.postRepo.GetPostsByUserID(user.UserID)
		if err == nil {
			users[j].Posts = posts
		} else if err != gorm.ErrRecordNotFound {
			log.WithError(err).Error("GetUsers.s.2")
		}
	}

	httpResponse.Message = users
	httpStatusCode = http.StatusOK
	return
}

// GetUser returns a user with the given userID and their posts.
func (s *UserService) GetUser(userID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	user, err := s.userRepo.GetUser(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "user not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("GetUser.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// fetch posts for the user
	posts, err := s.postRepo.GetPostsByUserID(user.UserID)
	if err == nil {
		user.Posts = posts
	} else if err != gorm.ErrRecordNotFound {
		log.WithError(err).Error("GetUser.s.2")
	}

	httpResponse.Message = user
	httpStatusCode = http.StatusOK
	return
}

// GetUserByAuthID retrieves a user by their authID and their posts.
func (s *UserService) GetUserByAuthID(authID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	user, err := s.userRepo.GetUserByAuthID(authID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "user not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("GetUserByAuthID.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// fetch posts for the user
	posts, err := s.postRepo.GetPostsByUserID(user.UserID)
	if err == nil {
		user.Posts = posts
	} else if err != gorm.ErrRecordNotFound {
		log.WithError(err).Error("GetUserByAuthID.s.2")
	}

	httpResponse.Message = user
	httpStatusCode = http.StatusOK
	return
}

// CreateUser adds a new user.
func (s *UserService) CreateUser(user *model.User) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user profile already exists
	_, err := s.userRepo.GetUserByAuthID(user.IDAuth)
	if err == nil {
		httpResponse.Message = "user profile already exists"
		httpStatusCode = http.StatusConflict
		return
	}

	// create the user profile
	err = s.userRepo.CreateUser(user)
	if err != nil {
		log.WithError(err).Error("CreateUser.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = user
	httpStatusCode = http.StatusCreated
	return
}

// UpdateUser updates an existing user.
func (s *UserService) UpdateUser(user *model.User) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	existingUser, err := s.userRepo.GetUserByAuthID(user.IDAuth)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "user not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("UpdateUser.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// if no changes are made, return the existing user
	if existingUser.FirstName == user.FirstName &&
		existingUser.LastName == user.LastName {
		httpResponse.Message = existingUser
		httpStatusCode = http.StatusOK
		return
	}

	// update the user fields
	existingUser.FirstName = user.FirstName
	existingUser.LastName = user.LastName

	// update the user profile
	err = s.userRepo.UpdateUser(existingUser)
	if err != nil {
		log.WithError(err).Error("UpdateUser.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = existingUser
	httpStatusCode = http.StatusOK
	return
}

// DeleteUser deletes a user with the given authID
// and their associated posts and hobbies.
func (s *UserService) DeleteUser(authID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	user, err := s.userRepo.GetUserByAuthID(authID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			httpResponse.Message = "user not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("DeleteUser.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete the user's posts
	if err := s.postRepo.DeletePostsByUserID(user.UserID); err != nil {
		log.WithError(err).Error("DeleteUser.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete the user's hobbies
	if err := s.hobbyRepo.DeleteHobbiesFromUser(user); err != nil {
		log.WithError(err).Error("DeleteUser.s.3")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete the user profile
	if err := s.userRepo.DeleteUser(user.UserID); err != nil {
		log.WithError(err).Error("DeleteUser.s.4")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// start deleting all auth-related data
	db := gdb.GetDB()
	authRepo := repo.NewAuthRepo(db)
	tempEmailRepo := repo.NewTempEmailRepo(db)
	twoFARepo := repo.NewTwoFARepo(db)
	twoFABackupRepo := repo.NewTwoFABackupRepo(db)

	// delete all 2fa backup codes for the user
	if err := twoFABackupRepo.DeleteTwoFABackup(authID); err != nil {
		log.WithError(err).Error("DeleteUser.s.5")
	}

	// delete the 2fa record for the user
	if err := twoFARepo.DeleteTwoFA(authID); err != nil {
		log.WithError(err).Error("DeleteUser.s.6")
	}

	// delete the temporary email for the user
	if err := tempEmailRepo.DeleteTempEmail(authID); err != nil {
		log.WithError(err).Error("DeleteUser.s.7")
	}

	// delete the auth record for the user
	if err := authRepo.DeleteAuth(authID); err != nil {
		log.WithError(err).Error("DeleteUser.s.8")
	}

	httpResponse.Message = "user deleted successfully"
	httpStatusCode = http.StatusOK
	return
}
