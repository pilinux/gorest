package model

import (
	"time"

	"gorm.io/gorm"
)

// TwoFA model - 'two_fas' table
type TwoFA struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	KeyMain   string
	KeyBackup string
	Status    string
	IDAuth    uint64
}

// Secret2FA - save encoded secrets in RAM temporarily
type Secret2FA struct {
	PassSHA []byte
	Secret  []byte
	Image   string
}

// InMemorySecret2FA - keep secrets temporarily
// in memory to setup 2FA
var InMemorySecret2FA = make(map[uint64]Secret2FA)
