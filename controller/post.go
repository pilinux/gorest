package controller

import (
	"net/http"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib/middleware"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetPosts - GET /posts
func GetPosts(c *gin.Context) {
	db := database.GetDB()
	posts := []model.Post{}

	if err := db.Find(&posts).Error; err != nil {
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

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err != nil {
		render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	c.ShouldBindJSON(&post)
	post.IDUser = user.UserID

	tx := db.Begin()
	if err := tx.Create(&post).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1201")
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		render(c, post, http.StatusCreated)
	}
}

// UpdatePost - PUT /posts/:id
func UpdatePost(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	post := model.Post{}
	id := c.Params.ByName("id")

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err != nil {
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&post).Error; err != nil {
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	c.ShouldBindJSON(&post)

	tx := db.Begin()
	if err := tx.Save(&post).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1211")
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		render(c, post, http.StatusOK)
	}
}

// DeletePost - DELETE /posts/:id
func DeletePost(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	post := model.Post{}
	id := c.Params.ByName("id")

	user.IDAuth = middleware.AuthID

	if err := db.Where("id_auth = ?", user.IDAuth).First(&user).Error; err != nil {
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&post).Error; err != nil {
		render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
		return
	}

	c.ShouldBindJSON(&post)

	tx := db.Begin()
	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1221")
		render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		render(c, gin.H{"Post ID# " + id: "Deleted!"}, http.StatusOK)
	}
}
