package model

import (
	"time"
)

// Post model - `posts` table
type Post struct {
	PostID    uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index" json:"-"`
	Title     string     `json:"Title,omitempty"`
	Body      string     `json:"Body,omitempty"`
	IDUser    uint       `json:"-"`
}
