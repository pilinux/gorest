package model

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/lib"
)

// Auth model - `auths` table
type Auth struct {
	AuthID    uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Email     string         `json:"Email"`
	Password  string         `json:"Password"`
	User      User           `gorm:"foreignkey:IDAuth;references:AuthID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
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

	// check password length
	// if more checks are required i.e. password pattern,
	// add all conditions here
	if len(aux.Password) < 6 {
		return errors.New("short password")
	}

	v.AuthID = aux.AuthID
	v.Email = aux.Email

	config := lib.HashPassConfig{
		Memory:      config.Security().HashPass.Memory,
		Iterations:  config.Security().HashPass.Iterations,
		Parallelism: config.Security().HashPass.Parallelism,
		SaltLength:  config.Security().HashPass.SaltLength,
		KeyLength:   config.Security().HashPass.KeyLength,
	}
	pass, err := lib.HashPass(config, aux.Password)
	if err != nil {
		return err
	}
	v.Password = pass

	return nil
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
