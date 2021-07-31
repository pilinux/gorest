package controller

import (
	"fmt"
	"net/http"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib/middleware"

	"github.com/gin-gonic/gin"
)

// GetPosts - GET /posts
func GetPosts(c *gin.Context) {
	db := database.GetDB()
	posts := []model.Post{}

	if err := db.Find(&posts).Error; err != nil {
		fmt.Println(err)
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		render(c, posts, http.StatusOK)
	}
}

// GetPost - GET /posts/:id
func GetPost(c *gin.Context) {
	db := database.GetDB()
	post := model.Post{}
	id := c.Params.ByName("id")

	if err := db.Where("post_id = ? ", id).First(&post).Error; err != nil {
		fmt.Println(err)
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		render(c, post, http.StatusOK)
	}
}

// CreatePost - POST /posts
func CreatePost(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	post := model.Post{}
	createPost := 0 // default

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err != nil {
		createPost = 1 // user data is not registered, so no post can be associated
		render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	if createPost == 0 {
		c.ShouldBindJSON(&post)
		post.IDUser = user.UserID

		tx := db.Begin()
		if err := tx.Create(&post).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		} else {
			tx.Commit()
			render(c, post, http.StatusCreated)
		}
	}
}

// UpdatePost - PUT /posts/:id
func UpdatePost(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	post := model.Post{}
	id := c.Params.ByName("id")
	updatePost := 0 // default

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err != nil {
		updatePost = 1 // user data is not registered, nothing can be updated
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	if updatePost == 0 {
		if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&post).Error; err != nil {
			updatePost = 1 // this post does not exist, or the user doesn't have any write access to this post
			render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
			return
		}
	}

	if updatePost == 0 {
		c.ShouldBindJSON(&post)

		tx := db.Begin()
		if err := tx.Save(&post).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		} else {
			tx.Commit()
			render(c, post, http.StatusOK)
		}
	}
}

// DeletePost - DELETE /posts/:id
func DeletePost(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	post := model.Post{}
	id := c.Params.ByName("id")
	deletePost := 0 // default

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err != nil {
		deletePost = 1 // user data is not registered, nothing can be deleted
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	if deletePost == 0 {
		if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&post).Error; err != nil {
			deletePost = 1 // this post does not exist, or the user doesn't have any write access to this post
			render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
			return
		}
	}

	if deletePost == 0 {
		c.ShouldBindJSON(&post)

		tx := db.Begin()
		if err := tx.Delete(&post).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
		} else {
			tx.Commit()
			render(c, gin.H{"Post ID# " + id: "Deleted!"}, http.StatusOK)
		}
	}
}
