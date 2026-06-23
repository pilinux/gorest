package handler

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	gdatabase "github.com/pilinux/gorest/database"
	gmodel "github.com/pilinux/gorest/database/model"

	"github.com/pilinux/gorest/example/database/model"
)

// GetHobbies retrieves all hobbies.
func GetHobbies() (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	db := gdatabase.GetDB()
	if db == nil {
		log.Error("error code: 1250: database connection not initialized")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}
	hobbies := []model.Hobby{}

	if err := db.Find(&hobbies).Error; err != nil {
		log.WithError(err).Error("error code: 1251")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if len(hobbies) == 0 {
		httpResponse.Message = "no hobby found"
		httpStatusCode = http.StatusNotFound
		return
	}

	httpResponse.Message = hobbies
	httpStatusCode = http.StatusOK
	return
}
