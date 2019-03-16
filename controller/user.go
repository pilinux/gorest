package controller

import (
	"fmt"

	"github.com/GoREST/database"
	"github.com/GoREST/database/model"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

var db *gorm.DB
var err error

// User struct alias
type User = model.User

// GetUsers: GET /users
func GetUsers(c *gin.Context) {
	db = database.GetDB()
	var users []User
	var posts []Post

	if err := db.Find(&users).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(404)
	} else {
		j := 0
		for _, user := range users {
			db.Table("posts").Joins("JOIN user_posts ON posts.id=user_posts.post_id").Joins("JOIN users ON users.id=user_posts.user_id").Where("users.id = ?", user.ID).Scan(&posts)
			//db.Raw("SELECT * FROM posts p JOIN user_posts up ON p.id = up.post_id JOIN users u ON u.id = up.user_id WHERE u.id = ?", user.ID).Scan(&posts)
			users[j].Posts = posts
			j++
		}
		c.JSON(200, users)
	}
}

// GetUser: GET /users/:id
func GetUser(c *gin.Context) {
	db = database.GetDB()
	id := c.Params.ByName("id")
	var user User
	var posts []Post

	if err := db.Where("id = ? ", id).First(&user).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(404)
	} else {
		db.Table("posts").Joins("JOIN user_posts ON posts.id=user_posts.post_id").Joins("JOIN users ON users.id=user_posts.user_id").Where("users.id = ?", id).Scan(&posts)
		//db.Raw("SELECT * FROM posts p JOIN user_posts up ON p.id = up.post_id JOIN users u ON u.id = up.user_id WHERE u.id = ?", id).Scan(&posts)
		user.Posts = posts
		c.JSON(200, user)
	}
}

// CreateUser: POST /users
func CreateUser(c *gin.Context) {
	db = database.GetDB()
	var user User

	c.BindJSON(&user)

	tx := db.Begin()
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		fmt.Println(err)
		c.AbortWithStatus(404)
	} else {
		tx.Commit()
		c.JSON(200, user)
	}
}

// UpdateUser: PUT /users/:id
func UpdateUser(c *gin.Context) {
	db = database.GetDB()
	var user User
	id := c.Params.ByName("id")

	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(404)
	}

	c.BindJSON(&user)

	tx := db.Begin()
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		fmt.Println(err)
		c.AbortWithStatus(501)
	} else {
		tx.Commit()
		c.JSON(200, user)
	}
}

// DeleteUser: DELETE /users/:id
func DeleteUser(c *gin.Context) {
	db = database.GetDB()
	id := c.Params.ByName("id")
	var user User
	var posts []Post

	if err := db.Where("id = ? ", id).Find(&user).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(404)
	} else {
		if err := db.Table("posts").Joins("JOIN user_posts ON posts.id=user_posts.post_id").Joins("JOIN users ON users.id=user_posts.user_id").Where("users.id = ?", id).Scan(&posts).Error; err != nil {
			fmt.Println(err)
			c.AbortWithStatus(404)
		} else {
			tx := db.Begin()

			for _, post := range posts {
				if err := tx.Delete(&post).Error; err != nil {
					tx.Rollback()
					fmt.Println(err)
					c.AbortWithStatus(404)
				}
			}

			if err := tx.Where("id = ? ", id).Delete(&user).Error; err != nil {
				tx.Rollback()
				fmt.Println(err)
				c.AbortWithStatus(404)
			} else {
				tx.Commit()
				c.JSON(200, gin.H{"id#" + id: "deleted"})
			}
		}
	}
}
