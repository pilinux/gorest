package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
)

// User model - `users` table
type User struct {
	gorm.Model
	Name     string `json:"Name"`
	Email    string `json:"Email"`
	Password string `json:"password"`
	Posts    []Post `gorm:"many2many:user_posts;foreignkey:UserID;association_foreignkey:ID"`
}

// UnmarshalJSON ...
func (u *User) UnmarshalJSON(b []byte) error {
	aux := struct {
		ID       uint   `json:"id"`
		Name     string `json:"Name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Posts    []Post `json:"posts"`
	}{}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	u.ID = aux.ID
	u.Name = aux.Name
	u.Email = aux.Email
	u.Posts = aux.Posts
	u.Password = HashPass(aux.Password)

	return nil
}

// HashPass ...
func HashPass(pass string) string {
	h := sha256.New()
	h.Write([]byte(pass))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MarshalJSON ...
func (u User) MarshalJSON() ([]byte, error) {
	aux := struct {
		ID    uint   `json:"id"`
		Name  string `json:"Name"`
		Email string `json:"email"`
		Posts []Post `json:"posts"`
	}{
		ID:    u.ID,
		Name:  u.Name,
		Email: u.Email,
		Posts: u.Posts,
	}

	return json.Marshal(aux)
}
