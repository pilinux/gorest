package repo

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/pilinux/gorest/example2/internal/database/model"
)

// HobbyRepo provides methods for hobby-related database operations.
type HobbyRepo struct {
	db *gorm.DB
}

// NewHobbyRepo returns a new HobbyRepo with the given database connection.
func NewHobbyRepo(conn *gorm.DB) *HobbyRepo {
	return &HobbyRepo{
		db: conn,
	}
}

// HobbyRepository defines the contract for hobby-related operations.
type HobbyRepository interface {
	GetHobbies(ctx context.Context) ([]model.Hobby, error)
	GetHobbiesByUserID(ctx context.Context, userID uint64) ([]model.Hobby, error)
	GetHobby(ctx context.Context, hobbyID uint64) (*model.Hobby, error)
	AddHobbyToUser(ctx context.Context, hobby *model.Hobby, user *model.User) error
	DeleteHobbyFromUser(ctx context.Context, hobbyID uint64, user *model.User) error
	DeleteHobbiesFromUser(ctx context.Context, user *model.User) error
}

// Compile-time check:
var _ HobbyRepository = (*HobbyRepo)(nil)

// GetHobbies retrieves all hobbies from the database.
func (r *HobbyRepo) GetHobbies(ctx context.Context) ([]model.Hobby, error) {
	var hobbies []model.Hobby
	if err := r.db.WithContext(ctx).Find(&hobbies).Error; err != nil {
		return nil, err
	}
	return hobbies, nil
}

// GetHobbiesByUserID retrieves hobbies for a user by userID.
func (r *HobbyRepo) GetHobbiesByUserID(ctx context.Context, userID uint64) ([]model.Hobby, error) {
	if userID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	var user model.User
	if err := r.db.WithContext(ctx).Preload("Hobbies").First(&user, userID).Error; err != nil {
		return nil, err
	}
	if len(user.Hobbies) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return user.Hobbies, nil
}

// GetHobby retrieves a hobby by its ID.
func (r *HobbyRepo) GetHobby(ctx context.Context, hobbyID uint64) (*model.Hobby, error) {
	var hobby model.Hobby
	if hobbyID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.WithContext(ctx).Where("hobby_id = ?", hobbyID).First(&hobby).Error; err != nil {
		return nil, err
	}
	return &hobby, nil
}

// AddHobbyToUser adds a hobby to an existing user.
func (r *HobbyRepo) AddHobbyToUser(ctx context.Context, hobby *model.Hobby, user *model.User) error {
	tNow := time.Now()
	hobby.HobbyID = 0 // auto-increment
	hobby.CreatedAt = tNow.Unix()
	hobby.UpdatedAt = tNow.Unix()

	// create the hobby if it doesn't exist
	if err := r.db.WithContext(ctx).FirstOrCreate(hobby, model.Hobby{Hobby: hobby.Hobby}).Error; err != nil {
		return err
	}

	// associate the hobby with the user
	return r.db.WithContext(ctx).Model(user).Association("Hobbies").Append(hobby)
}

// DeleteHobbyFromUser removes a hobby from a user by hobbyID.
func (r *HobbyRepo) DeleteHobbyFromUser(ctx context.Context, hobbyID uint64, user *model.User) error {
	hobby, err := r.GetHobby(ctx, hobbyID)
	if err != nil {
		return err
	}

	// remove the hobby from the user's hobbies
	if err := r.db.WithContext(ctx).Model(user).Association("Hobbies").Delete(hobby); err != nil {
		return err
	}

	// check if the hobby has no other associations
	count := r.db.WithContext(ctx).Model(hobby).Association("Users").Count()
	if count > 0 {
		// if the hobby is still associated with other users, exit here
		return nil
	}

	// delete the hobby as it has no other associations
	return r.db.WithContext(ctx).Where("hobby_id = ?", hobbyID).Delete(&model.Hobby{}).Error
}

// DeleteHobbiesFromUser removes all hobbies from a user.
func (r *HobbyRepo) DeleteHobbiesFromUser(ctx context.Context, user *model.User) error {
	if user == nil || user.UserID == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// preload hobbies for the user
		if err := tx.Preload("Hobbies").First(user, user.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil // no hobbies to delete
			}
			return err
		}
		if len(user.Hobbies) == 0 {
			return nil // no hobbies to delete
		}

		// copy the hobbies to a new slice to avoid modifying the original slice during iteration
		hobbiesCopy := make([]model.Hobby, len(user.Hobbies))
		copy(hobbiesCopy, user.Hobbies)

		// remove all associations
		if err := tx.Model(user).Association("Hobbies").Clear(); err != nil {
			return err
		}

		// delete orphaned hobbies
		for _, hobby := range hobbiesCopy {
			count := tx.Model(&hobby).Association("Users").Count()
			if count == 0 {
				if err := tx.Where("hobby_id = ?", hobby.HobbyID).Delete(&model.Hobby{}).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})
}
