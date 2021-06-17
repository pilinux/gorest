package model

import (
	"time"

	"gorm.io/gorm"
)

// Post model - `posts` table
type Post struct {
	PostID    uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Title     string         `json:"Title,omitempty"`
	Body      string         `json:"Body,omitempty"`
	IDUser    uint64         `json:"-"`
}
