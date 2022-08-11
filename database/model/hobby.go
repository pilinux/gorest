package model

import (
	"time"

	"gorm.io/gorm"
)

// Hobby model - `hobbies` table
type Hobby struct {
	HobbyID   uint64         `gorm:"primaryKey" json:"hobbyID,omitempty"`
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Hobby     string         `json:"hobby,omitempty"`
	Users     []User         `gorm:"many2many:user_hobbies" json:"-"`
}
