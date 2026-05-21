package model_test

import (
	"testing"

	"github.com/pilinux/gorest/database/model"
)

func TestSecret2FAStoreSetClonesInput(t *testing.T) {
	store := model.NewSecret2FAStore()
	value := model.Secret2FA{
		PassHash: []byte{1, 2, 3},
		KeySalt:  []byte{4, 5, 6},
		Secret:   []byte{7, 8, 9},
		Image:    "qr.png",
	}

	store.Set(1, value)
	value.PassHash[0] = 10
	value.KeySalt[0] = 11
	value.Secret[0] = 12

	got, ok := store.Get(1)
	if !ok {
		t.Fatal("expected stored value")
	}
	if got.PassHash[0] != 1 {
		t.Fatalf("PassHash was mutated through input aliasing: got %d, want 1", got.PassHash[0])
	}
	if got.KeySalt[0] != 4 {
		t.Fatalf("KeySalt was mutated through input aliasing: got %d, want 4", got.KeySalt[0])
	}
	if got.Secret[0] != 7 {
		t.Fatalf("Secret was mutated through input aliasing: got %d, want 7", got.Secret[0])
	}
}

func TestSecret2FAStoreGetReturnsClone(t *testing.T) {
	store := model.NewSecret2FAStore()
	store.Set(1, model.Secret2FA{
		PassHash: []byte{1, 2, 3},
		KeySalt:  []byte{4, 5, 6},
		Secret:   []byte{7, 8, 9},
		Image:    "qr.png",
	})

	got, ok := store.Get(1)
	if !ok {
		t.Fatal("expected stored value")
	}
	got.PassHash[0] = 10
	got.KeySalt[0] = 11
	got.Secret[0] = 12

	again, ok := store.Get(1)
	if !ok {
		t.Fatal("expected stored value")
	}
	if again.PassHash[0] != 1 {
		t.Fatalf("PassHash was mutated through Get aliasing: got %d, want 1", again.PassHash[0])
	}
	if again.KeySalt[0] != 4 {
		t.Fatalf("KeySalt was mutated through Get aliasing: got %d, want 4", again.KeySalt[0])
	}
	if again.Secret[0] != 7 {
		t.Fatalf("Secret was mutated through Get aliasing: got %d, want 7", again.Secret[0])
	}
}

func TestSecret2FAStoreDelete(t *testing.T) {
	store := model.NewSecret2FAStore()
	store.Set(1, model.Secret2FA{Image: "qr.png"})

	store.Delete(1)

	if _, ok := store.Get(1); ok {
		t.Fatal("expected deleted value to be absent")
	}
}
