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
	var posts []model.Post

	if err := db.Where("email = ? ", email).Find(&user).Error; err != nil {
		fmt.Println(err)
		return nil, err
	}

	db.Table("posts").Joins("JOIN user_posts ON posts.id=user_posts.post_id").Joins("JOIN users ON users.id=user_posts.user_id").Where("users.id = ?", user.ID).Scan(&posts)
	//db.Raw("SELECT * FROM posts p JOIN user_posts up ON p.id = up.post_id JOIN users u ON u.id = up.user_id WHERE u.id = ?", user.ID).Scan(&posts)
	user.Posts = posts

	return &user, nil
}
