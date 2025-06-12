// Package repo provides data access and persistence logic for the application's models.
// It defines repository types and methods that interact with the database.
package repo

import (
	"context"
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

// UserRepository defines the contract for user-related operations.
type UserRepository interface {
	GetUsers(ctx context.Context) ([]model.User, error)
	GetUser(ctx context.Context, userID uint64) (*model.User, error)
	GetUserByAuthID(ctx context.Context, authID uint64) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUser(ctx context.Context, user *model.User) error
	DeleteUser(ctx context.Context, userID uint64) error
	DeleteUserByAuthID(ctx context.Context, authID uint64) error
}

// Compile-time check:
var _ UserRepository = (*UserRepo)(nil)

// GetUsers returns all users from the database.
func (r *UserRepo) GetUsers(ctx context.Context) ([]model.User, error) {
	var users []model.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUser returns a user with the given userID from the database.
func (r *UserRepo) GetUser(ctx context.Context, userID uint64) (*model.User, error) {
	var user model.User
	if userID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByAuthID returns a user with the given authID from the database.
func (r *UserRepo) GetUserByAuthID(ctx context.Context, authID uint64) (*model.User, error) {
	var user model.User
	if authID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	if err := r.db.WithContext(ctx).Where("id_auth = ?", authID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a new user in the database.
func (r *UserRepo) CreateUser(ctx context.Context, user *model.User) error {
	tNow := time.Now()
	user.UserID = 0 // auto-increment
	user.CreatedAt = tNow.Unix()
	user.UpdatedAt = tNow.Unix()
	user.Hobbies = nil // clear hobbies to avoid inserting them unintentionally
	return r.db.WithContext(ctx).Create(user).Error
}

// UpdateUser updates an existing user in the database.
func (r *UserRepo) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now().Unix()
	user.Hobbies = nil // clear hobbies to avoid updating them unintentionally
	return r.db.WithContext(ctx).Save(user).Error
}

// DeleteUser deletes a user with the given userID from the database.
func (r *UserRepo) DeleteUser(ctx context.Context, userID uint64) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&model.User{}).Error
}

// DeleteUserByAuthID deletes a user with the given authID from the database.
func (r *UserRepo) DeleteUserByAuthID(ctx context.Context, authID uint64) error {
	return r.db.WithContext(ctx).Where("id_auth = ?", authID).Delete(&model.User{}).Error
}
