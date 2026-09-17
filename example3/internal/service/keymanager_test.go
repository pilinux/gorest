package service

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/pilinux/crypt"
	"github.com/pilinux/crypt/envelope"
	"github.com/pilinux/gorest/example3/internal/database/model"
)

const (
	currentSecret = "0123456789abcdef0123456789abcdef01234567"
	oldSecret     = "old-0123456789abcdef0123456789abcdef0123"
)

// storedWith builds a SecretKey document whose wrapped master key was sealed
// with the KEK derived from secret.
func storedWith(t *testing.T, secret string, masterKey []byte) *model.SecretKey {
	t.Helper()
	kek, err := scheme.DeriveKEK(secret)
	if err != nil {
		t.Fatalf("DeriveKEK error: %v", err)
	}
	wrapped, err := envelope.WrapKey(kek, masterKey)
	if err != nil {
		t.Fatalf("WrapKey error: %v", err)
	}
	return &model.SecretKey{
		ID:         bson.NewObjectID(),
		KeyName:    model.MasterKeyName,
		WrappedKey: wrapped,
	}
}

func TestSetSecrets(t *testing.T) {
	t.Run("validNoRotation", func(t *testing.T) {
		km := NewKeyManager()
		if err := km.SetSecrets(currentSecret, ""); err != nil {
			t.Fatalf("SetSecrets error: %v", err)
		}
		if km.IsRotationConfigured() {
			t.Error("rotation should not be configured without an old secret")
		}
	})

	t.Run("shortCurrent", func(t *testing.T) {
		km := NewKeyManager()
		if err := km.SetSecrets("short", ""); !errors.Is(err, envelope.ErrSecretTooShort) {
			t.Errorf("err = %v, want ErrSecretTooShort", err)
		}
	})

	t.Run("shortOld", func(t *testing.T) {
		km := NewKeyManager()
		if err := km.SetSecrets(currentSecret, "short"); err == nil {
			t.Error("expected error for short old secret")
		}
	})

	t.Run("oldEqualsCurrentNoRotation", func(t *testing.T) {
		km := NewKeyManager()
		if err := km.SetSecrets(currentSecret, currentSecret); err != nil {
			t.Fatalf("SetSecrets error: %v", err)
		}
		if km.IsRotationConfigured() {
			t.Error("rotation must be off when old == current (nothing to rotate)")
		}
	})

	t.Run("distinctOldConfiguresRotation", func(t *testing.T) {
		km := NewKeyManager()
		if err := km.SetSecrets(currentSecret, oldSecret); err != nil {
			t.Fatalf("SetSecrets error: %v", err)
		}
		if !km.IsRotationConfigured() {
			t.Error("rotation should be configured with a distinct old secret")
		}
	})
}

func TestEnsureAndLoad_Bootstrap(t *testing.T) {
	km := NewKeyManager()
	if err := km.SetSecrets(currentSecret, ""); err != nil {
		t.Fatalf("SetSecrets error: %v", err)
	}

	store := &fakeKeyStore{} // fresh, empty database
	if err := km.EnsureAndLoad(context.Background(), store); err != nil {
		t.Fatalf("EnsureAndLoad error: %v", err)
	}

	mk, err := km.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey error: %v", err)
	}
	if len(mk) != envelope.KeySize {
		t.Errorf("master key len = %d, want %d", len(mk), envelope.KeySize)
	}
	if store.updateCalls != 0 {
		t.Error("bootstrap must not trigger a rotation update")
	}

	// the stored wrapped key must unwrap to the loaded master key
	kek, _ := scheme.DeriveKEK(currentSecret)
	unwrapped, err := envelope.UnwrapKey(kek, store.stored.WrappedKey)
	if err != nil {
		t.Fatalf("stored key does not unwrap with current KEK: %v", err)
	}
	if !bytes.Equal(unwrapped, mk) {
		t.Error("stored key does not match loaded master key")
	}
}

func TestEnsureAndLoad_ExistingCurrentSecret(t *testing.T) {
	masterKey, _ := envelope.GenerateMasterKey()
	store := &fakeKeyStore{stored: storedWith(t, currentSecret, masterKey)}

	km := NewKeyManager()
	_ = km.SetSecrets(currentSecret, "")
	if err := km.EnsureAndLoad(context.Background(), store); err != nil {
		t.Fatalf("EnsureAndLoad error: %v", err)
	}

	mk, _ := km.MasterKey()
	if !bytes.Equal(mk, masterKey) {
		t.Error("loaded master key does not match the stored one")
	}
	if store.updateCalls != 0 {
		t.Error("no rotation should happen when the current secret already fits")
	}
}

func TestEnsureAndLoad_Rotation(t *testing.T) {
	masterKey, _ := envelope.GenerateMasterKey()
	// the database still holds a key wrapped with the OLD secret
	store := &fakeKeyStore{stored: storedWith(t, oldSecret, masterKey)}

	km := NewKeyManager()
	_ = km.SetSecrets(currentSecret, oldSecret)
	if err := km.EnsureAndLoad(context.Background(), store); err != nil {
		t.Fatalf("EnsureAndLoad error: %v", err)
	}

	mk, _ := km.MasterKey()
	if !bytes.Equal(mk, masterKey) {
		t.Error("rotation changed the master key, it must stay the same")
	}
	if store.updateCalls != 1 {
		t.Errorf("update calls = %d, want 1 (re-wrap during rotation)", store.updateCalls)
	}
	// the KEKs only wrap and unwrap the master key, so neither outlives the load
	if km.kek != nil || km.kekOld != nil {
		t.Error("KEKs were kept after the master key was loaded")
	}

	// after rotation the stored key must unwrap with the CURRENT secret
	kek, _ := scheme.DeriveKEK(currentSecret)
	unwrapped, err := envelope.UnwrapKey(kek, store.stored.WrappedKey)
	if err != nil {
		t.Fatalf("re-wrapped key does not unwrap with current KEK: %v", err)
	}
	if !bytes.Equal(unwrapped, masterKey) {
		t.Error("re-wrapped key does not match the master key")
	}
}

func TestEnsureAndLoad_WrongSecretNoOld(t *testing.T) {
	masterKey, _ := envelope.GenerateMasterKey()
	store := &fakeKeyStore{stored: storedWith(t, oldSecret, masterKey)}

	km := NewKeyManager()
	_ = km.SetSecrets(currentSecret, "") // no old secret to fall back on
	err := km.EnsureAndLoad(context.Background(), store)
	if !errors.Is(err, ErrUndecryptableMasterKey) {
		t.Errorf("err = %v, want ErrUndecryptableMasterKey", err)
	}
	// a changed secret is an authentication failure, and says so
	if !errors.Is(err, envelope.ErrEnvelopeAuth) {
		t.Errorf("err = %v, want it to wrap envelope.ErrEnvelopeAuth", err)
	}
}

// TestEnsureAndLoad_NonAuthFailureSkipsRotation: only an authentication failure
// means the secret changed. A stored value that authenticates under the current
// KEK but is not a master key is refused as it is, not handed to the old secret.
func TestEnsureAndLoad_NonAuthFailureSkipsRotation(t *testing.T) {
	kek, err := scheme.DeriveKEK(currentSecret)
	if err != nil {
		t.Fatalf("DeriveKEK error: %v", err)
	}
	// sealed under the current KEK exactly as WrapKey seals, but half a key long
	short, err := crypt.EncryptByteXChacha20poly1305WithNonceAppended(kek, make([]byte, envelope.KeySize/2))
	if err != nil {
		t.Fatalf("encrypt error: %v", err)
	}
	store := &fakeKeyStore{stored: &model.SecretKey{
		ID:         bson.NewObjectID(),
		KeyName:    model.MasterKeyName,
		WrappedKey: short,
	}}

	km := NewKeyManager()
	if err := km.SetSecrets(currentSecret, oldSecret); err != nil {
		t.Fatalf("SetSecrets error: %v", err)
	}
	err = km.EnsureAndLoad(context.Background(), store)
	if !errors.Is(err, ErrUndecryptableMasterKey) || !errors.Is(err, envelope.ErrInvalidKeySize) {
		t.Errorf("err = %v, want ErrUndecryptableMasterKey wrapping envelope.ErrInvalidKeySize", err)
	}
	if store.updateCalls != 0 {
		t.Errorf("update calls = %d, want 0 (no rotation for a non-auth failure)", store.updateCalls)
	}
}

func TestEnsureAndLoad_SecretNotSet(t *testing.T) {
	km := NewKeyManager()
	err := km.EnsureAndLoad(context.Background(), &fakeKeyStore{})
	if !errors.Is(err, ErrSecretNotSet) {
		t.Errorf("err = %v, want ErrSecretNotSet", err)
	}
}

// TestKeyManagerConcurrentInit: the key lifecycle is serialized, so overlapping
// calls never wipe a KEK another call is still using. Meant to run with -race.
func TestKeyManagerConcurrentInit(t *testing.T) {
	masterKey, _ := envelope.GenerateMasterKey()
	store := &fakeKeyStore{stored: storedWith(t, currentSecret, masterKey)}

	km := NewKeyManager()
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if err := km.SetSecrets(currentSecret, oldSecret); err != nil {
				t.Errorf("SetSecrets error: %v", err)
				return
			}
			km.IsRotationConfigured()
			if err := km.EnsureAndLoad(context.Background(), store); err != nil {
				t.Errorf("EnsureAndLoad error: %v", err)
				return
			}
			if _, err := km.MasterKey(); err != nil {
				t.Errorf("MasterKey error: %v", err)
			}
		})
	}
	wg.Wait()

	mk, err := km.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey error: %v", err)
	}
	if !bytes.Equal(mk, masterKey) {
		t.Error("concurrent init did not load the stored master key")
	}
	// every goroutine ends on EnsureAndLoad, so the last one wiped the KEKs
	if km.kek != nil || km.kekOld != nil {
		t.Error("KEKs were kept after the master key was loaded")
	}
	if store.updateCalls != 0 {
		t.Errorf("update calls = %d, want 0 (the stored key already fits)", store.updateCalls)
	}
}

func TestMasterKey_NotLoaded(t *testing.T) {
	km := NewKeyManager()
	if _, err := km.MasterKey(); !errors.Is(err, ErrMasterKeyNotLoaded) {
		t.Errorf("err = %v, want ErrMasterKeyNotLoaded", err)
	}
}
