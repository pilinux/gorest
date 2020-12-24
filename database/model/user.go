package model

import (
	"time"
)

// User model - `users` table
type User struct {
	UserID    uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	FirstName string     `json:"FirstName"`
	LastName  string     `json:"LastName"`
	IDAuth    uint
	Posts     []Post  `gorm:"foreignkey:IDPost;association_foreignkey:UserID"`
	Hobbies   []Hobby `gorm:"many2many:user_hobbies"`
}
