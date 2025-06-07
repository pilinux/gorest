package model

import (
	"errors"
	"strings"
)

// Post model - `posts` table
type Post struct {
	PostID    uint64 `gorm:"primaryKey" json:"postID,omitempty" structs:"postID,omitempty"`
	CreatedAt int64  `json:"createdAt,omitempty" structs:"createdAt,omitempty"`
	UpdatedAt int64  `json:"updatedAt,omitempty" structs:"updatedAt,omitempty"`
	Title     string `json:"title,omitempty" structs:"title,omitempty"`
	Body      string `json:"body,omitempty" structs:"body,omitempty"`
	IDUser    uint64 `gorm:"index" json:"-"`
}

// Trim trims leading and trailing spaces from Title and Body,
// and validates that none of the fields are empty.
func (p *Post) Trim() error {
	p.Title = strings.TrimSpace(p.Title)
	p.Body = strings.TrimSpace(p.Body)

	if p.Title == "" {
		return errors.New("title is required")
	}
	if p.Body == "" {
		return errors.New("body is required")
	}

	return nil
}
