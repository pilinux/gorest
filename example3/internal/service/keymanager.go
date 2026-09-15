// Package service holds example3's business logic: the master key, and
// encrypting and decrypting text, numbers and files.
package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pilinux/crypt/envelope"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/pilinux/gorest/example3/internal/database/model"
	"github.com/pilinux/gorest/example3/internal/repo"
)

// Errors returned by KeyManager.
var (
	// ErrSecretNotSet means the env secret has not been set.
	ErrSecretNotSet = errors.New("service: encryption secret is not set")

	// ErrMasterKeyNotLoaded means the master key was asked for before
	// EnsureAndLoad succeeded.
	ErrMasterKeyNotLoaded = errors.New("service: master key is not loaded")

	// ErrUndecryptableMasterKey means neither the current nor the old env secret
	// can decrypt the stored master key.
	ErrUndecryptableMasterKey = errors.New("service: master key cannot be decrypted with the current or old secret")
)

// KeyManager holds the keys at runtime. It derives KEKs from the env secrets,
// makes sure the database has a wrapped master key, re-wraps it when the secret
// changes, and keeps the master key in memory.
//
// KEKs only wrap the master key, so they are wiped once it is loaded.
type KeyManager struct {
	mu        sync.RWMutex
	kek       []byte // current KEK, from ENCRYPTION_SECRET; nil once loaded
	kekOld    []byte // previous KEK, only when rotating; nil otherwise
	masterKey []byte // plaintext master key from the database
}

// NewKeyManager returns an empty KeyManager. Call SetSecrets and then
// EnsureAndLoad before using MasterKey.
func NewKeyManager() *KeyManager {
	return &KeyManager{}
}

// SetSecrets derives the current KEK from current, and the old KEK from old.
//
// old only counts when it is set and differs from current. Otherwise there is
// nothing to rotate, and no old KEK is kept.
func (km *KeyManager) SetSecrets(current, old string) error {
	kek, err := scheme.DeriveKEK(current)
	if err != nil {
		return err
	}

	var kekOld []byte
	if old != "" && old != current {
		kekOld, err = scheme.DeriveKEK(old)
		if err != nil {
			envelope.Zero(kek)
			return fmt.Errorf("old secret: %w", err)
		}
	}

	km.mu.Lock()
	km.kek = kek
	km.kekOld = kekOld
	km.mu.Unlock()
	return nil
}

// IsRotationConfigured reports whether an old secret, different from the
// current one, was given. It only means something until EnsureAndLoad wipes
// the KEKs.
func (km *KeyManager) IsRotationConfigured() bool {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return len(km.kekOld) > 0
}

// EnsureAndLoad loads the master key from the database into memory, creating
// one on a fresh database. If the current KEK cannot unwrap it but the old one
// can, it is re-wrapped under the current KEK; the master key itself never
// changes.
//
// Call SetSecrets first. The KEKs are wiped once the key is loaded.
func (km *KeyManager) EnsureAndLoad(ctx context.Context, store repo.KeyStore) error {
	km.mu.RLock()
	kek := km.kek
	kekOld := km.kekOld
	km.mu.RUnlock()

	if len(kek) == 0 {
		return ErrSecretNotSet
	}

	// a new master key, saved only if the database has none yet. The key used
	// is always the one read back below, so this plaintext is wiped right away.
	candidateKey, err := envelope.GenerateMasterKey()
	if err != nil {
		return err
	}
	wrapped, err := envelope.WrapKey(kek, candidateKey)
	envelope.Zero(candidateKey)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	candidate := &model.SecretKey{
		ID:         bson.NewObjectID(),
		KeyName:    model.MasterKeyName,
		WrappedKey: wrapped,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	stored, err := store.EnsureMasterKey(ctx, candidate)
	if err != nil {
		return err
	}

	plain, err := unwrapStored(ctx, store, stored, kek, kekOld)
	if err != nil {
		return err
	}

	km.mu.Lock()
	km.masterKey = plain
	envelope.Zero(km.kek)
	envelope.Zero(km.kekOld)
	km.kek, km.kekOld = nil, nil
	km.mu.Unlock()
	return nil
}

// unwrapStored decrypts the stored master key. If it is still wrapped under
// kekOld, it is re-wrapped under kek and saved.
func unwrapStored(ctx context.Context, store repo.KeyStore, stored *model.SecretKey, kek, kekOld []byte) ([]byte, error) {
	// try the current secret first
	plain, err := envelope.UnwrapKey(kek, stored.WrappedKey)
	if err == nil {
		return plain, nil
	}

	// a changed secret only ever shows up as ErrEnvelopeAuth. Any other error,
	// or no old secret, is not something the old secret can fix.
	if !errors.Is(err, envelope.ErrEnvelopeAuth) || len(kekOld) == 0 {
		return nil, fmt.Errorf("%w: %w", ErrUndecryptableMasterKey, err)
	}

	// rotation: recover with the old secret, then re-wrap with the current one
	plain, err = envelope.UnwrapKey(kekOld, stored.WrappedKey)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUndecryptableMasterKey, err)
	}
	rewrapped, err := envelope.WrapKey(kek, plain)
	if err == nil {
		err = store.UpdateWrappedKey(ctx, stored.ID, rewrapped)
	}
	if err != nil {
		envelope.Zero(plain)
		return nil, err
	}
	return plain, nil
}

// MasterKey returns the loaded master key, or an error if it is not loaded yet.
func (km *KeyManager) MasterKey() ([]byte, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()
	if len(km.masterKey) == 0 {
		return nil, ErrMasterKeyNotLoaded
	}
	return km.masterKey, nil
}
