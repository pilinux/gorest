package controller

import (
	"fmt"
	"net/http"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib/middleware"

	"github.com/gin-gonic/gin"
)

// GetUsers - GET /users
func GetUsers(c *gin.Context) {
	db := database.GetDB()
	users := []model.User{}
	posts := []model.Post{}
	hobbies := []model.Hobby{}

	if err := db.Find(&users).Error; err != nil {
		fmt.Println(err)
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		j := 0
		for _, user := range users {
			db.Model(&posts).Where("id_user = ?", user.UserID).Find(&posts)
			users[j].Posts = posts
			db.Model(&hobbies).Joins("JOIN user_hobbies ON user_hobbies.hobby_hobby_id=hobbies.hobby_id").
				Joins("JOIN users ON users.user_id=user_hobbies.user_user_id").
				Where("users.user_id = ?", user.UserID).
				Find(&hobbies)
			users[j].Hobbies = hobbies
			j++
		}
		render(c, users, http.StatusOK)
	}
}

// GetUser - GET /users/:id
func GetUser(c *gin.Context) {
	db := database.GetDB()
	id := c.Params.ByName("id")
	user := model.User{}
	posts := []model.Post{}
	hobbies := []model.Hobby{}

	if err := db.Where("user_id = ? ", id).First(&user).Error; err != nil {
		fmt.Println(err)
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		db.Model(&posts).Where("id_user = ?", id).Find(&posts)
		user.Posts = posts
		db.Model(&hobbies).Joins("JOIN user_hobbies ON user_hobbies.hobby_hobby_id=hobbies.hobby_id").
			Joins("JOIN users ON users.user_id=user_hobbies.user_user_id").
			Where("users.user_id = ?", user.UserID).
			Find(&hobbies)
		user.Hobbies = hobbies
		render(c, user, http.StatusOK)
	}
}

// CreateUser - POST /users
func CreateUser(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	createUser := 0 // default

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err == nil {
		createUser = 1 // user data already registered
		render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	if createUser == 0 {
		c.ShouldBindJSON(&user)

		tx := db.Begin()
		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		} else {
			tx.Commit()
			render(c, user, http.StatusCreated)
		}
	}
}

// UpdateUser - PUT /users
func UpdateUser(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	updateUser := 0 // default

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err != nil {
		updateUser = 1 // user data is not registered, nothing can be updated
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	if updateUser == 0 {
		c.ShouldBindJSON(&user)
		fmt.Println(user.UserID)

		tx := db.Begin()
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		} else {
			tx.Commit()
			render(c, user, http.StatusOK)
		}
	}
}

// AddHobby - PUT /users/hobbies
func AddHobby(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	hobby := model.Hobby{}
	hobbyFound := 0 // default = 0, do not proceed = 1, create new hobby = 2

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err != nil {
		hobbyFound = 1 // user data is not registered, do not proceed
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	c.ShouldBindJSON(&hobby)

	if err := db.First(&hobby, "hobby = ?", hobby.Hobby).Error; err != nil {
		hobbyFound = 2 // create new hobby
	}

	if hobbyFound == 2 {
		tx := db.Begin()
		if err := tx.Create(&hobby).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		} else {
			tx.Commit()
			hobbyFound = 0
		}
	}

	if hobbyFound == 0 {
		user.Hobbies = append(user.Hobbies, hobby)
		tx := db.Begin()
		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		} else {
			tx.Commit()
			render(c, user, http.StatusOK)
		}
	}
}
