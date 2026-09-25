package service

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/pilinux/crypt/envelope"
	"github.com/pilinux/gorest/example3/internal/database/model"
)

// loadedKeys returns a stub master-key provider holding a fresh random key.
func loadedKeys(t *testing.T) stubKeys {
	t.Helper()
	key, err := envelope.GenerateMasterKey()
	if err != nil {
		t.Fatalf("GenerateMasterKey error: %v", err)
	}
	return stubKeys{key: key}
}

func TestTextCrypt_EncryptDecryptText(t *testing.T) {
	svc := NewTextCryptService(loadedKeys(t))
	const plaintext = "the eagle lands at midnight"

	resp, code := svc.EncryptText(plaintext)
	if code != http.StatusOK {
		t.Fatalf("EncryptText status = %d, want 200", code)
	}
	cipher, ok := resp.Message.(model.CiphertextResponse)
	if !ok {
		t.Fatalf("unexpected message type %T", resp.Message)
	}
	if cipher.Ciphertext == "" {
		t.Fatal("empty ciphertext")
	}

	resp, code = svc.DecryptText(cipher.Ciphertext)
	if code != http.StatusOK {
		t.Fatalf("DecryptText status = %d, want 200", code)
	}
	plain, ok := resp.Message.(model.PlaintextResponse)
	if !ok {
		t.Fatalf("unexpected message type %T", resp.Message)
	}
	if plain.Plaintext != plaintext {
		t.Errorf("round-trip mismatch: got %q, want %q", plain.Plaintext, plaintext)
	}
}

func TestTextCrypt_EncryptText_Empty(t *testing.T) {
	svc := NewTextCryptService(loadedKeys(t))
	if _, code := svc.EncryptText(""); code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", code)
	}
}

func TestTextCrypt_EncryptText_TooLarge(t *testing.T) {
	svc := NewTextCryptService(loadedKeys(t))
	if _, code := svc.EncryptText(strings.Repeat("a", maxTextLen+1)); code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", code)
	}
}

func TestTextCrypt_DecryptText_Errors(t *testing.T) {
	svc := NewTextCryptService(loadedKeys(t))

	if _, code := svc.DecryptText(""); code != http.StatusBadRequest {
		t.Errorf("empty token status = %d, want 400", code)
	}
	if _, code := svc.DecryptText("not-a-valid-token"); code != http.StatusBadRequest {
		t.Errorf("bad token status = %d, want 400", code)
	}
}

func TestTextCrypt_NumberRoundTrip(t *testing.T) {
	svc := NewTextCryptService(loadedKeys(t))
	n := int64(-42)

	resp, code := svc.EncryptNumber(&n)
	if code != http.StatusOK {
		t.Fatalf("EncryptNumber status = %d, want 200", code)
	}
	cipher := resp.Message.(model.CiphertextResponse)

	resp, code = svc.DecryptNumber(cipher.Ciphertext)
	if code != http.StatusOK {
		t.Fatalf("DecryptNumber status = %d, want 200", code)
	}
	num := resp.Message.(model.NumberResponse)
	if num.Number != n {
		t.Errorf("round-trip mismatch: got %d, want %d", num.Number, n)
	}
}

func TestTextCrypt_Number_Errors(t *testing.T) {
	svc := NewTextCryptService(loadedKeys(t))

	if _, code := svc.EncryptNumber(nil); code != http.StatusBadRequest {
		t.Errorf("nil number status = %d, want 400", code)
	}
	if _, code := svc.DecryptNumber(""); code != http.StatusBadRequest {
		t.Errorf("empty token status = %d, want 400", code)
	}
	if _, code := svc.DecryptNumber("garbage"); code != http.StatusBadRequest {
		t.Errorf("bad token status = %d, want 400", code)
	}
}

func TestTextCrypt_MasterKeyError(t *testing.T) {
	svc := NewTextCryptService(stubKeys{err: errors.New("not loaded")})

	if _, code := svc.EncryptText("x"); code != http.StatusInternalServerError {
		t.Errorf("EncryptText status = %d, want 500", code)
	}
	n := int64(1)
	if _, code := svc.EncryptNumber(&n); code != http.StatusInternalServerError {
		t.Errorf("EncryptNumber status = %d, want 500", code)
	}
}

// TestTextCrypt_TokenTypesDontMix - a number token doesn't open as text, and an
// 8-byte text token doesn't open as a number.
func TestTextCrypt_TokenTypesDontMix(t *testing.T) {
	svc := NewTextCryptService(loadedKeys(t))

	number := int64(0x4142434445464748) // "ABCDEFGH" as bytes
	resp, code := svc.EncryptNumber(&number)
	if code != http.StatusOK {
		t.Fatalf("EncryptNumber status = %d, want 200", code)
	}
	numberToken := resp.Message.(model.CiphertextResponse).Ciphertext

	resp, code = svc.EncryptText("ABCDEFGH")
	if code != http.StatusOK {
		t.Fatalf("EncryptText status = %d, want 200", code)
	}
	textToken := resp.Message.(model.CiphertextResponse).Ciphertext

	tests := []struct {
		name    string
		decrypt func(string) (int, any)
		token   string
	}{
		{
			name: "number token as text",
			decrypt: func(tok string) (int, any) {
				resp, code := svc.DecryptText(tok)
				return code, resp.Message
			},
			token: numberToken,
		},
		{
			name: "text token as number",
			decrypt: func(tok string) (int, any) {
				resp, code := svc.DecryptNumber(tok)
				return code, resp.Message
			},
			token: textToken,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if code, msg := tc.decrypt(tc.token); code != http.StatusBadRequest {
				t.Errorf("status = %d (%v), want 400", code, msg)
			}
		})
	}
}
