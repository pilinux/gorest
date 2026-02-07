// Package model contains all the models required
// for a functional database management system.
package model

import (
	"errors"
	"strings"
)

// User represents a user in the users table.
type User struct {
	UserID    uint64  `gorm:"primaryKey" json:"userID,omitempty"`
	CreatedAt int64   `json:"createdAt,omitempty"`
	UpdatedAt int64   `json:"updatedAt,omitempty"`
	FirstName string  `json:"firstName,omitempty"`
	LastName  string  `json:"lastName,omitempty"`
	IDAuth    uint64  `gorm:"index" json:"-"`
	Posts     []Post  `gorm:"-" json:"posts,omitempty"`
	Hobbies   []Hobby `gorm:"many2many:user_hobbies" json:"hobbies,omitempty"`
}

// Trim trims leading and trailing spaces from FirstName and LastName,
// and validates that none of the fields are empty.
func (u *User) Trim() error {
	u.FirstName = strings.TrimSpace(u.FirstName)
	u.LastName = strings.TrimSpace(u.LastName)

	if u.FirstName == "" {
		return errors.New("firstName is required")
	}
	if u.LastName == "" {
		return errors.New("lastName is required")
	}

	return nil
}
