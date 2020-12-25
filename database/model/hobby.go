package model

import (
	"time"
)

// Hobby model - `hobbies` table
type Hobby struct {
	HobbyID   uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index" json:"-"`
	Hobby     string     `json:"Hobby,omitempty"`
	Users     []User     `gorm:"many2many:user_hobbies" json:"-"`
}
