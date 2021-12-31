package model

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/alexedwards/argon2id"

	"github.com/pilinux/gorest/config"
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
	if v.Password = HashPass(aux.Password); v.Password == "error" {
		return errors.New("HashPass failed")
	}

	return nil
}

// HashPass ...
func HashPass(pass string) string {
	configureHash := config.Security().HashPass
	params := &argon2id.Params{
		Memory:      configureHash.Memory * 1024, // the amount of memory used by the Argon2 algorithm (in kibibytes)
		Iterations:  configureHash.Iterations,    // the number of iterations (or passes) over the memory
		Parallelism: configureHash.Parallelism,   // the number of threads (or lanes) used by the algorithm
		SaltLength:  configureHash.SaltLength,    // length of the random salt. 16 bytes is recommended for password hashing
		KeyLength:   configureHash.KeyLength,     // length of the generated key (or password hash). 16 bytes or more is recommended
	}
	h, err := argon2id.CreateHash(pass, params)
	if err != nil {
		return "error"
	}
	return h
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
