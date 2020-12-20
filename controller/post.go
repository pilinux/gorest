package controller

import (
	"fmt"
	"net/http"

	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"
	"github.com/piLinux/GoREST/lib/middleware"

	"github.com/gin-gonic/gin"
)

// GetPosts - GET /posts
func GetPosts(c *gin.Context) {
	db := database.GetDB()
	posts := []model.Post{}

	if err := db.Find(&posts).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, posts)
	}
}

// GetPost - GET /posts/:id
func GetPost(c *gin.Context) {
	db := database.GetDB()
	post := model.Post{}
	id := c.Params.ByName("id")

	if err := db.Where("post_id = ? ", id).First(&post).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusOK, post)
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
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if createPost == 0 {
		c.ShouldBindJSON(&post)
		post.IDUser = user.UserID

		tx := db.Begin()
		if err := tx.Create(&post).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			tx.Commit()
			c.JSON(http.StatusCreated, post)
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
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if updatePost == 0 {
		if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&post).Error; err != nil {
			updatePost = 1 // this post does not exist, or the user doesn't have any write access to this post
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	}

	if updatePost == 0 {
		c.ShouldBindJSON(&post)

		tx := db.Begin()
		if err := tx.Save(&post).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			tx.Commit()
			c.JSON(http.StatusOK, post)
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
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	if deletePost == 0 {
		if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&post).Error; err != nil {
			deletePost = 1 // this post does not exist, or the user doesn't have any write access to this post
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	}

	if deletePost == 0 {
		c.ShouldBindJSON(&post)

		tx := db.Begin()
		if err := tx.Delete(&post).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		} else {
			tx.Commit()
			c.JSON(http.StatusOK, gin.H{"Post ID# " + id: "Deleted!"})
		}
	}
}
