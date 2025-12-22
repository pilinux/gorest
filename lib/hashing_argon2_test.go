package lib_test

import (
	"encoding/base64"
	"testing"

	"github.com/pilinux/gorest/lib"
	"github.com/pilinux/gorest/service"
)

func TestGetArgon2Key(t *testing.T) {
	password := "my-secret-password"
	// 1. generate salt
	salt, err := service.RandomByte(16)
	if err != nil {
		t.Fatalf("failed to generate salt: %v", err)
	}

	// 2. derive key
	key := lib.GetArgon2Key([]byte(password), salt, 32)
	if len(key) != 32 {
		t.Errorf("expected key length 32, got %d", len(key))
	}

	// 3. encrypt data
	data := []byte("some secret data")
	encrypted, err := lib.Encrypt(data, key)
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// 4. simulate saving/loading salt (base64)
	saltStr := base64.StdEncoding.EncodeToString(salt)
	decodedSalt, err := base64.StdEncoding.DecodeString(saltStr)
	if err != nil {
		t.Fatalf("salt decoding failed: %v", err)
	}

	// 5. derive key again
	key2 := lib.GetArgon2Key([]byte(password), decodedSalt, 32)

	// 6. decrypt
	decrypted, err := lib.Decrypt(encrypted, key2)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	if string(decrypted) != string(data) {
		t.Errorf("decrypted data mismatch. got %s, want %s", string(decrypted), string(data))
	}
}
