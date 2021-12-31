package service

import (
	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
)

// GetUserByEmail ...
func GetUserByEmail(email string) (*model.Auth, error) {
	db := database.GetDB()

	var auth model.Auth

	if err := db.Where("email = ? ", email).First(&auth).Error; err != nil {
		return nil, err
	}

	return &auth, nil
}
