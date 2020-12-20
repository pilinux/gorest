package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// Auth model - `auths` table
type Auth struct {
	AuthID    uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	Email     string     `json:"Email"`
	Password  string     `json:"Password"`
	User      User       `gorm:"foreignkey:IDAuth;association_foreignkey:AuthID"`
}

// UnmarshalJSON ...
func (v *Auth) UnmarshalJSON(b []byte) error {
	aux := struct {
		AuthID   uint   `json:"AuthID"`
		Email    string `json:"Email"`
		Password string `json:"Password"`
	}{}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	v.AuthID = aux.AuthID
	v.Email = aux.Email
	v.Password = HashPass(aux.Password)

	return nil
}

// HashPass ...
func HashPass(pass string) string {
	h := sha256.New()
	h.Write([]byte(pass))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// MarshalJSON ...
func (v Auth) MarshalJSON() ([]byte, error) {
	aux := struct {
		AuthID uint   `json:"AuthId"`
		Email  string `json:"Email"`
	}{
		AuthID: v.AuthID,
		Email:  v.Email,
	}

	return json.Marshal(aux)
}
