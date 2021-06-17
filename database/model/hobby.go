package model

import (
	"time"

	"gorm.io/gorm"
)

// Hobby model - `hobbies` table
type Hobby struct {
	HobbyID   uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Hobby     string         `json:"Hobby,omitempty"`
	Users     []User         `gorm:"many2many:user_hobbies" json:"-"`
}
