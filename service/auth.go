package service

import (
	"fmt"

	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"
)

// GetUserByEmail ...
func GetUserByEmail(email string) (*model.Auth, error) {
	db := database.GetDB()

	var auth model.Auth

	if err := db.Where("email = ? ", email).Find(&auth).Error; err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &auth, nil
}
