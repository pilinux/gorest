package model

import (
	"time"

	"gorm.io/gorm"
)

// Post model - `posts` table
type Post struct {
	PostID    uint64         `gorm:"primaryKey" json:"postID,omitempty" structs:"postID,omitempty"`
	CreatedAt time.Time      `json:"createdAt,omitempty" structs:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty" structs:"updatedAt,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Title     string         `json:"title,omitempty" structs:"title,omitempty"`
	Body      string         `json:"body,omitempty" structs:"body,omitempty"`
	IDUser    uint64         `json:"-"`
}
