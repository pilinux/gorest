package controller

import (
	"net/http"
	"time"

	"github.com/pilinux/gorest/database"
	"github.com/pilinux/gorest/database/model"
	"github.com/pilinux/gorest/lib/middleware"
	"github.com/pilinux/gorest/lib/renderer"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetPosts - GET /posts
func GetPosts(c *gin.Context) {
	db := database.GetDB()
	posts := []model.Post{}

	if err := db.Find(&posts).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "not found"}, http.StatusNotFound)
	} else {
		renderer.Render(c, posts, http.StatusOK)
	}
}

// GetPost - GET /posts/:id
func GetPost(c *gin.Context) {
	db := database.GetDB()
	post := model.Post{}
	id := c.Params.ByName("id")
	errorMsg := model.ErrorMsg{}

	if err := db.Where("post_id = ? ", id).First(&post).Error; err != nil {
		errorMsg.HTTPCode = http.StatusNotFound
		errorMsg.Message = "not found"
		renderer.Render(c, errorMsg, http.StatusNotFound, "error.html")
	} else {
		renderer.Render(c, post, http.StatusOK, "read-article.html")
	}
}

// CreatePost - POST /posts
func CreatePost(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	post := model.Post{}
	postFinal := model.Post{}

	userIDAuth := middleware.AuthID

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&user).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "no user profile found"}, http.StatusForbidden)
		return
	}

	// bind JSON
	if err := c.ShouldBindJSON(&post); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// user must not be able to manipulate all fields
	postFinal.Title = post.Title
	postFinal.Body = post.Body
	postFinal.IDUser = user.UserID

	tx := db.Begin()
	if err := tx.Create(&postFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1201")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		renderer.Render(c, postFinal, http.StatusCreated)
	}
}

// UpdatePost - PUT /posts/:id
func UpdatePost(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	post := model.Post{}
	postFinal := model.Post{}
	id := c.Params.ByName("id")

	userIDAuth := middleware.AuthID

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&user).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "no user profile found"}, http.StatusForbidden)
		return
	}

	// does the post exist + does user have right to modify this post
	if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&postFinal).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "operation not possible"}, http.StatusForbidden)
		return
	}

	// bind JSON
	if err := c.ShouldBindJSON(&post); err != nil {
		renderer.Render(c, gin.H{"msg": "bad request"}, http.StatusBadRequest)
		return
	}

	// user must not be able to manipulate all fields
	postFinal.UpdatedAt = time.Now()
	postFinal.Title = post.Title
	postFinal.Body = post.Body

	tx := db.Begin()
	if err := tx.Save(&postFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1211")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		renderer.Render(c, postFinal, http.StatusOK)
	}
}

// DeletePost - DELETE /posts/:id
func DeletePost(c *gin.Context) {
	db := database.GetDB()
	user := model.User{}
	post := model.Post{}
	id := c.Params.ByName("id")

	userIDAuth := middleware.AuthID

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&user).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "no user profile found"}, http.StatusForbidden)
		return
	}

	// does the post exist + does user have right to delete this post
	if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&post).Error; err != nil {
		renderer.Render(c, gin.H{"msg": "operation not possible"}, http.StatusForbidden)
		return
	}

	tx := db.Begin()
	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1221")
		renderer.Render(c, gin.H{"msg": "internal server error"}, http.StatusInternalServerError)
	} else {
		tx.Commit()
		renderer.Render(c, gin.H{"Post ID# " + id: "Deleted!"}, http.StatusOK)
	}
}
