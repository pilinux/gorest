package lib_test

import (
	"bytes"
	"testing"

	"github.com/pilinux/gorest/lib"
)

func TestEncryptDecrypt(t *testing.T) {
	// test with valid key
	key := []byte("1234567890123456")
	plaintext := []byte("hello world")

	// test encryption
	ciphertext, err := lib.Encrypt(plaintext, key)
	if err != nil {
		t.Errorf("error encrypting: %v", err)
	}

	// test decryption
	decrypted, err := lib.Decrypt(ciphertext, key)
	if err != nil {
		t.Errorf("error decrypting: %v", err)
	}

	// check that plaintext and decrypted text are equal
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("plaintext and decrypted text do not match")
	}

	// test decryption with incorrect key
	wrongKey := []byte("0123456789012345")
	_, err = lib.Decrypt(ciphertext, wrongKey)
	if err == nil {
		t.Error("expected error when decrypting with incorrect key, but got nil")
	}

	// test decryption with invalid ciphertext
	invalidCiphertext := []byte("invalid ciphertext")
	_, err = lib.Decrypt(invalidCiphertext, key)
	if err == nil {
		t.Error("expected error when decrypting invalid ciphertext, but got nil")
	}

	// test encryption with invalid key (size must be 16, 24 or 32 bytes)
	invalidKey := []byte("short key")
	_, err = lib.Encrypt(plaintext, invalidKey)
	if err == nil {
		t.Error("expected error when creating cipher with invalid key, but got nil")
	}

	// test encryption with nil key
	_, err = lib.Encrypt(plaintext, nil)
	if err == nil {
		t.Error("expected error when creating cipher with nil key, but got nil")
	}

	// test decryption with invalid key (size must be 16, 24 or 32 bytes)
	_, err = lib.Decrypt(ciphertext, invalidKey)
	if err == nil {
		t.Error("expected error when creating cipher with invalid key, but got nil")
	}

	// test decryption with nil key
	_, err = lib.Decrypt(ciphertext, nil)
	if err == nil {
		t.Error("expected error when creating cipher with nil key, but got nil")
	}
}
