package model

import (
	"errors"
	"strings"
)

// Hobby represents a hobby in the hobbies table.
type Hobby struct {
	HobbyID   uint64 `gorm:"primaryKey" json:"hobbyID,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty"`
	Hobby     string `json:"hobby,omitempty"`
	Users     []User `gorm:"many2many:user_hobbies" json:"-"`
}

// Trim trims leading and trailing spaces from Hobby,
// and validates that Hobby is not empty.
func (h *Hobby) Trim() error {
	h.Hobby = strings.TrimSpace(h.Hobby)

	if h.Hobby == "" {
		return errors.New("hobby is required")
	}

	return nil
}
