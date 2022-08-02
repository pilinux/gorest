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
