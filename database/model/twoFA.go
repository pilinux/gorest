package model

import (
	"sync"
	"time"

	"gorm.io/gorm"
)

// TwoFA represents the two_fas table.
type TwoFA struct {
	ID        uint64         `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	KeyMain   string         `json:"-"`
	KeyBackup string         `json:"-"`
	KeySalt   string         `json:"-"`
	UUIDSHA   string         `json:"-"`
	UUIDEnc   string         `json:"-"`
	Status    string         `json:"-"`
	IDAuth    uint64         `gorm:"index" json:"-"`
}

// TwoFABackup represents the two_fa_backups table.
type TwoFABackup struct {
	ID        uint64    `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time `json:"-"`
	Code      string    `gorm:"-" json:"code"`
	CodeHash  string    `json:"-"`
	IDAuth    uint64    `gorm:"index" json:"-"`
}

// Secret2FA holds encoded secrets temporarily in RAM.
type Secret2FA struct {
	PassHash []byte `json:"-"`
	KeySalt  []byte `json:"-"`
	Secret   []byte `json:"-"`
	Image    string `json:"-"`
}

// Secret2FAStore provides thread-safe access to in-memory 2FA secrets.
type Secret2FAStore struct {
	mu   sync.RWMutex
	data map[uint64]Secret2FA
}

// NewSecret2FAStore creates a new Secret2FAStore.
func NewSecret2FAStore() *Secret2FAStore {
	return &Secret2FAStore{
		data: make(map[uint64]Secret2FA),
	}
}

// Get retrieves a Secret2FA from the store.
func (s *Secret2FAStore) Get(key uint64) (Secret2FA, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// Set stores a Secret2FA in the store.
func (s *Secret2FAStore) Set(key uint64, value Secret2FA) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete removes a Secret2FA from the store.
func (s *Secret2FAStore) Delete(key uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

// InMemorySecret2FA keeps secrets temporarily
// in memory to set up 2FA.
var InMemorySecret2FA = NewSecret2FAStore()
