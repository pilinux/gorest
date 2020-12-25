package model

import (
	"time"
)

// User model - `users` table
type User struct {
	UserID    uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index" json:"-"`
	FirstName string     `json:"FirstName,omitempty"`
	LastName  string     `json:"LastName,omitempty"`
	IDAuth    uint       `json:"-"`
	Posts     []Post     `gorm:"foreignkey:IDPost;association_foreignkey:UserID" json:",omitempty"`
	Hobbies   []Hobby    `gorm:"many2many:user_hobbies" json:",omitempty"`
}
