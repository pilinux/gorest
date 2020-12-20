package model

import (
	"time"
)

// Post model - `posts` table
type Post struct {
	PostID    uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	Title     string     `json:"Title"`
	Body      string     `json:"Body"`
	IDUser    uint
}
