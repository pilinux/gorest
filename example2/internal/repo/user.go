// Package repo provides data access and persistence logic for the application's models.
// It defines repository types and methods that interact with the database.
package repo

import (
	"time"

	"gorm.io/gorm"

	"github.com/pilinux/gorest/example2/internal/database/model"
)

// UserRepo provides methods for user-related database operations.
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo returns a new UserRepo with the given database connection.
func NewUserRepo(conn *gorm.DB) *UserRepo {
	return &UserRepo{
		db: conn,
	}
}

// GetUsers returns all users from the database.
func (r *UserRepo) GetUsers() ([]model.User, error) {
	var users []model.User
	if err := r.db.Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUser returns a user with the given userID from the database.
func (r *UserRepo) GetUser(userID uint64) (*model.User, error) {
	var user model.User
	if userID == 0 {
		return &model.User{}, gorm.ErrRecordNotFound
	}
	if err := r.db.Where("user_id = ?", userID).First(&user).Error; err != nil {
		return &model.User{}, err
	}
	return &user, nil
}

// GetUserByAuthID returns a user with the given authID from the database.
func (r *UserRepo) GetUserByAuthID(authID uint64) (*model.User, error) {
	var user model.User
	if authID == 0 {
		return &model.User{}, gorm.ErrRecordNotFound
	}
	if err := r.db.Where("id_auth = ?", authID).First(&user).Error; err != nil {
		return &model.User{}, err
	}
	return &user, nil
}

// CreateUser creates a new user in the database.
func (r *UserRepo) CreateUser(user *model.User) error {
	tNow := time.Now()
	user.UserID = 0 // auto-increment
	user.CreatedAt = tNow.Unix()
	user.UpdatedAt = tNow.Unix()
	user.Hobbies = nil // clear hobbies to avoid inserting them unintentionally
	return r.db.Create(user).Error
}

// UpdateUser updates an existing user in the database.
func (r *UserRepo) UpdateUser(user *model.User) error {
	user.UpdatedAt = time.Now().Unix()
	user.Hobbies = nil // clear hobbies to avoid updating them unintentionally
	return r.db.Save(user).Error
}

// DeleteUser deletes a user with the given userID from the database.
func (r *UserRepo) DeleteUser(userID uint64) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.User{}).Error
}

// DeleteUserByAuthID deletes a user with the given authID from the database.
func (r *UserRepo) DeleteUserByAuthID(authID uint64) error {
	return r.db.Where("id_auth = ?", authID).Delete(&model.User{}).Error
}
