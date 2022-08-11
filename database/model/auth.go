// Package model contains all the models required
// for a functional database management system
package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/pilinux/gorest/config"
	"github.com/pilinux/gorest/lib"
)

// Auth model - `auths` table
type Auth struct {
	AuthID    uint64         `gorm:"primaryKey" json:"authID,omitempty"`
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Email     string         `json:"email"`
	Password  string         `json:"password"`
}

// UnmarshalJSON ...
func (v *Auth) UnmarshalJSON(b []byte) error {
	aux := struct {
		AuthID   uint64 `json:"authID"`
		Email    string `json:"email"`
		Password string `json:"password"`
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
	v.Email = strings.TrimSpace(aux.Email)

	config := lib.HashPassConfig{
		Memory:      config.SecurityConfigAll.HashPass.Memory,
		Iterations:  config.SecurityConfigAll.HashPass.Iterations,
		Parallelism: config.SecurityConfigAll.HashPass.Parallelism,
		SaltLength:  config.SecurityConfigAll.HashPass.SaltLength,
		KeyLength:   config.SecurityConfigAll.HashPass.KeyLength,
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
		AuthID uint64 `json:"authID"`
		Email  string `json:"email"`
	}{
		AuthID: v.AuthID,
		Email:  strings.TrimSpace(v.Email),
	}

	return json.Marshal(aux)
}
