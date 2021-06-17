package model

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Auth model - `auths` table
type Auth struct {
	AuthID    uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Email     string         `json:"Email"`
	Password  string         `json:"Password"`
	Users     User           `gorm:"foreignkey:IDAuth;references:AuthID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// UnmarshalJSON ...
func (v *Auth) UnmarshalJSON(b []byte) error {
	aux := struct {
		AuthID   uint64 `json:"AuthID"`
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
		AuthID uint64 `json:"AuthId"`
		Email  string `json:"Email"`
	}{
		AuthID: v.AuthID,
		Email:  v.Email,
	}

	return json.Marshal(aux)
}
