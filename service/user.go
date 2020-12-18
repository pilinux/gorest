package service

import (
	"fmt"

	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"
)

// GetUserByEmail ...
func GetUserByEmail(email string) (*model.User, error) {
	db := database.GetDB()

	var user model.User

	if err := db.Where("email = ? ", email).Find(&user).Error; err != nil {
		fmt.Println(err)
		return nil, err
	}

	return &user, nil
}
