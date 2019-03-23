package controller

import (
	"fmt"

	"github.com/piLinux/GoREST/database"
	"github.com/piLinux/GoREST/database/model"

	"github.com/gin-gonic/gin"
)

// Post struct alias
type Post = model.Post

// GetPosts - GET /posts
func GetPosts(c *gin.Context) {
	db := database.GetDB()
	var posts []Post

	if err := db.Find(&posts).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(404)
	} else {
		c.JSON(200, posts)
	}
}

// GetPost - GET /posts/:id
func GetPost(c *gin.Context) {
	db := database.GetDB()
	id := c.Params.ByName("id")
	var post Post

	if err := db.Where("id = ? ", id).First(&post).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(404)
	} else {
		c.JSON(200, post)
	}
}

// CreatePost - POST /posts
func CreatePost(c *gin.Context) {
	db := database.GetDB()
	var post Post

	c.BindJSON(&post)

	tx := db.Begin()
	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		fmt.Println(err)
		c.AbortWithStatus(404)
	} else {
		tx.Commit()
		c.JSON(200, post)
	}
}

// UpdatePost - PUT /posts/:id
func UpdatePost(c *gin.Context) {
	db := database.GetDB()
	var post Post
	id := c.Params.ByName("id")

	if err := db.Where("id = ?", id).First(&post).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(404)
	}

	c.BindJSON(&post)

	tx := db.Begin()
	if err := tx.Save(&post).Error; err != nil {
		tx.Rollback()
		fmt.Println(err)
		c.AbortWithStatus(501)
	} else {
		tx.Commit()
		c.JSON(200, post)
	}
}

// DeletePost - DELETE /posts/:id
func DeletePost(c *gin.Context) {
	db := database.GetDB()
	id := c.Params.ByName("id")
	var post Post

	if err := db.Where("id = ? ", id).Find(&post).Error; err != nil {
		fmt.Println(err)
		c.AbortWithStatus(404)
	} else {
		tx := db.Begin()

		if err := tx.Where("id = ? ", id).Delete(&post).Error; err != nil {
			tx.Rollback()
			fmt.Println(err)
			c.AbortWithStatus(404)
		} else {
			tx.Commit()
			c.JSON(200, gin.H{"id#" + id: "deleted"})
		}
	}
}
