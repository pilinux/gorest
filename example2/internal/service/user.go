// Package service provides business logic and application services for the API.
// It defines service types and methods that interact with repositories and implement core operations.
package service

import (
	"context"
	"errors"
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
func (s *UserService) GetUsers(ctx context.Context) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	users, err := s.userRepo.GetUsers(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

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

	// collect user IDs to batch-load related data
	userIDs := make([]uint64, len(users))
	for j := range users {
		userIDs[j] = users[j].UserID
	}

	// fetch posts for all users in a single query (avoids N+1)
	postsByUser, err := s.postRepo.GetPostsByUserIDs(ctx, userIDs)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}
		log.WithError(err).Error("GetUsers.s.2")
	} else {
		for j := range users {
			users[j].Posts = postsByUser[users[j].UserID]
		}
	}

	// fetch hobbies for all users in a single query (avoids N+1)
	hobbiesByUser, err := s.hobbyRepo.GetHobbiesByUserIDs(ctx, userIDs)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}
		log.WithError(err).Error("GetUsers.s.3")
	} else {
		for j := range users {
			users[j].Hobbies = hobbiesByUser[users[j].UserID]
		}
	}

	httpResponse.Message = users
	httpStatusCode = http.StatusOK
	return
}

// GetUser returns a user with the given userID and their posts.
func (s *UserService) GetUser(ctx context.Context, userID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	user, err := s.userRepo.GetUser(ctx, userID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	posts, err := s.postRepo.GetPostsByUserID(ctx, user.UserID)
	if err == nil {
		user.Posts = posts
	} else if errors.Is(err, context.Canceled) {
		httpResponse.Message = "request canceled"
		httpStatusCode = http.StatusRequestTimeout
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("GetUser.s.2")
	}

	// fetch hobbies for the user
	hobbies, err := s.hobbyRepo.GetHobbiesByUserID(ctx, user.UserID)
	if err == nil {
		user.Hobbies = hobbies
	} else if errors.Is(err, context.Canceled) {
		httpResponse.Message = "request canceled"
		httpStatusCode = http.StatusRequestTimeout
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("GetUser.s.3")
	}

	httpResponse.Message = user
	httpStatusCode = http.StatusOK
	return
}

// GetUserByAuthID retrieves a user by their authID and their posts.
func (s *UserService) GetUserByAuthID(ctx context.Context, authID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	user, err := s.userRepo.GetUserByAuthID(ctx, authID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	posts, err := s.postRepo.GetPostsByUserID(ctx, user.UserID)
	if err == nil {
		user.Posts = posts
	} else if errors.Is(err, context.Canceled) {
		httpResponse.Message = "request canceled"
		httpStatusCode = http.StatusRequestTimeout
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("GetUserByAuthID.s.2")
	}

	// fetch hobbies for the user
	hobbies, err := s.hobbyRepo.GetHobbiesByUserID(ctx, user.UserID)
	if err == nil {
		user.Hobbies = hobbies
	} else if errors.Is(err, context.Canceled) {
		httpResponse.Message = "request canceled"
		httpStatusCode = http.StatusRequestTimeout
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("GetUserByAuthID.s.3")
	}

	httpResponse.Message = user
	httpStatusCode = http.StatusOK
	return
}

// CreateUser adds a new user.
func (s *UserService) CreateUser(ctx context.Context, user *model.User) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user profile already exists
	_, err := s.userRepo.GetUserByAuthID(ctx, user.IDAuth)
	if err == nil {
		httpResponse.Message = "user profile already exists"
		httpStatusCode = http.StatusConflict
		return
	}
	if errors.Is(err, context.Canceled) {
		httpResponse.Message = "request canceled"
		httpStatusCode = http.StatusRequestTimeout
		return
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.WithError(err).Error("CreateUser.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// create the user profile
	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("CreateUser.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = user
	httpStatusCode = http.StatusCreated
	return
}

// UpdateUser updates an existing user.
func (s *UserService) UpdateUser(ctx context.Context, user *model.User) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	existingUser, err := s.userRepo.GetUserByAuthID(ctx, user.IDAuth)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
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
	err = s.userRepo.UpdateUser(ctx, existingUser)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

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
func (s *UserService) DeleteUser(ctx context.Context, authID uint64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	// check if the user exists
	user, err := s.userRepo.GetUserByAuthID(ctx, authID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			httpResponse.Message = "user not found"
			httpStatusCode = http.StatusNotFound
			return
		}

		log.WithError(err).Error("DeleteUser.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete the user profile and every associated record (posts, hobbies,
	// 2FA, 2FA backup codes, temp email, auth) in a single transaction so the
	// cascade is all-or-nothing: if any step fails, nothing is committed and no
	// orphaned/half-deleted rows remain.
	db := gdb.GetDB()
	if db == nil {
		log.Error("DeleteUser.s.0: database connection not initialized")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		postRepo := repo.NewPostRepo(tx)
		hobbyRepo := repo.NewHobbyRepo(tx)
		userRepo := repo.NewUserRepo(tx)
		authRepo := repo.NewAuthRepo(tx)
		tempEmailRepo := repo.NewTempEmailRepo(tx)
		twoFARepo := repo.NewTwoFARepo(tx)
		twoFABackupRepo := repo.NewTwoFABackupRepo(tx)

		// delete the user's posts
		if err := postRepo.DeletePostsByUserID(ctx, user.UserID); err != nil {
			return err
		}

		// delete the user's hobbies
		if err := hobbyRepo.DeleteHobbiesFromUser(ctx, user); err != nil {
			return err
		}

		// delete the user profile
		if err := userRepo.DeleteUser(ctx, user.UserID); err != nil {
			return err
		}

		// delete all 2fa backup codes for the user
		if err := twoFABackupRepo.DeleteTwoFABackup(ctx, authID); err != nil {
			return err
		}

		// delete the 2fa record for the user
		if err := twoFARepo.DeleteTwoFA(ctx, authID); err != nil {
			return err
		}

		// delete the temporary email for the user
		if err := tempEmailRepo.DeleteTempEmail(ctx, authID); err != nil {
			return err
		}

		// delete the auth record for the user
		if err := authRepo.DeleteAuth(ctx, authID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			httpResponse.Message = "request canceled"
			httpStatusCode = http.StatusRequestTimeout
			return
		}

		log.WithError(err).Error("DeleteUser.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = "user deleted successfully"
	httpStatusCode = http.StatusOK
	return
}
