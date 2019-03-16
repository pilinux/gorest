package model

import (
	"github.com/jinzhu/gorm"
)

// Post model - `posts` table
type Post struct {
	gorm.Model
	Title string `json:"Title"`
	Body  string `json:"Body"`
}
