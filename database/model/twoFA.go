package model

import (
	"time"

	"gorm.io/gorm"
)

// TwoFA model - 'two_fas' table
type TwoFA struct {
	ID        uint64         `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	KeyMain   string         `json:"-"`
	KeyBackup string         `json:"-"`
	UUIDSHA   string         `json:"-"`
	UUIDEnc   string         `json:"-"`
	Status    string         `json:"-"`
	IDAuth    uint64         `gorm:"index" json:"-"`
}

// TwoFABackup model - 'two_fa_backups' table
type TwoFABackup struct {
	ID        uint64    `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"-"`
	Code      string    `gorm:"-" json:"code"`
	CodeHash  string    `json:"-"`
	IDAuth    uint64    `gorm:"index" json:"-"`
}

// Secret2FA - save encoded secrets in RAM temporarily
type Secret2FA struct {
	PassSHA []byte `json:"-"`
	Secret  []byte `json:"-"`
	Image   string `json:"-"`
}

// InMemorySecret2FA - keep secrets temporarily
// in memory to setup 2FA
var InMemorySecret2FA = make(map[uint64]Secret2FA)
