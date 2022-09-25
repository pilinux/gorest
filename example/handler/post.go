package handler

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	gdatabase "github.com/pilinux/gorest/database"
	gmodel "github.com/pilinux/gorest/database/model"

	"github.com/pilinux/gorest/example/database/model"
)

// GetPosts handles jobs for controller.GetPosts
func GetPosts() (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	db := gdatabase.GetDB()
	posts := []model.Post{}

	if err := db.Find(&posts).Error; err != nil {
		log.WithError(err).Error("error code: 1201")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(posts) == 0 {
		httpResponse.Message = "no article found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = posts
	httpStatusCode = http.StatusOK
	return
}

// GetPost handles jobs for controller.GetPost
func GetPost(id string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	db := gdatabase.GetDB()
	post := model.Post{}

	if err := db.Where("post_id = ?", id).First(&post).Error; err != nil {
		httpResponse.Message = "article not found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = post
	httpStatusCode = http.StatusOK
	return
}

// CreatePost handles jobs for controller.CreatePost
func CreatePost(userIDAuth uint64, post model.Post) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	db := gdatabase.GetDB()
	user := model.User{}
	postFinal := model.Post{}

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&user).Error; err != nil {
		httpResponse.Message = "no user profile found"
		httpStatusCode = http.StatusForbidden
		return
	}

	// user must not be able to manipulate all fields
	postFinal.Title = post.Title
	postFinal.Body = post.Body
	postFinal.IDUser = user.UserID

	tx := db.Begin()
	if err := tx.Create(&postFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1211")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = postFinal
	httpStatusCode = http.StatusCreated
	return
}

// UpdatePost handles jobs for controller.UpdatePost
func UpdatePost(userIDAuth uint64, id string, post model.Post) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	db := gdatabase.GetDB()
	user := model.User{}
	postFinal := model.Post{}

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&user).Error; err != nil {
		httpResponse.Message = "no user profile found"
		httpStatusCode = http.StatusForbidden
		return
	}

	// does the post exist + does the user have right to modify this post
	if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&postFinal).Error; err != nil {
		httpResponse.Message = "user may not have access to perform this task"
		httpStatusCode = http.StatusForbidden
		return
	}

	// user must not be able to manipulate all fields
	postFinal.UpdatedAt = time.Now()
	postFinal.Title = post.Title
	postFinal.Body = post.Body

	tx := db.Begin()
	if err := tx.Save(&postFinal).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1221")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = postFinal
	httpStatusCode = http.StatusOK
	return
}

// DeletePost handles jobs for controller.DeletePost
func DeletePost(userIDAuth uint64, id string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	db := gdatabase.GetDB()
	user := model.User{}
	post := model.Post{}

	// does the user have an existing profile
	if err := db.Where("id_auth = ?", userIDAuth).First(&user).Error; err != nil {
		httpResponse.Message = "no user profile found"
		httpStatusCode = http.StatusForbidden
		return
	}

	// does the post exist + does the user have right to delete this post
	if err := db.Where("post_id = ?", id).Where("id_user = ?", user.UserID).First(&post).Error; err != nil {
		httpResponse.Message = "user may not have access to perform this task"
		httpStatusCode = http.StatusForbidden
		return
	}

	tx := db.Begin()
	if err := tx.Delete(&post).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("error code: 1231")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	tx.Commit()

	httpResponse.Message = "post ID# " + id + " deleted!"
	httpStatusCode = http.StatusOK
	return
}
