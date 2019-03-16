package model

import (
	"github.com/jinzhu/gorm"
)

type Post struct {
	gorm.Model
	Title string `json:"Title"`
	Body  string `json:"Body"`
}
