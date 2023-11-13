package lib_test

import (
	"crypto"
	"testing"

	"github.com/pilinux/twofactor"

	"github.com/pilinux/gorest/lib"
)

type totpTest struct {
	account string
	issuer  string
	hash    crypto.Hash
	digits  int
}

var tests = []totpTest{
	{
		account: "test@email.com",
		issuer:  "TestIssuer",
		hash:    crypto.SHA1,
		digits:  6,
	},
	{
		account: "test2@email.com",
		issuer:  "TestIssuer2",
		hash:    crypto.SHA256,
		digits:  8,
	},
	{
		account: "test3@email.com",
		issuer:  "TestIssuer3",
		hash:    crypto.SHA512,
		digits:  10,
	},
}

func TestTOTP(t *testing.T) {
	for _, test := range tests {
		// test: create new TOTP object
		otpByte, err := lib.NewTOTP(test.account, test.issuer, test.hash, test.digits)
		if err != nil {
			t.Fatalf("NewTOTP returned error: %v", err)
		}
		// check that the returned byte array is not empty
		if len(otpByte) == 0 {
			t.Errorf("NewTOTP returned empty byte array")
		}
		encryptedMessage := otpByte

		// test: encode QR in bytes
		qrByte, err := lib.NewQR(otpByte, test.issuer)
		if err != nil {
			t.Fatalf("NewQR returned error: %v", err)
		}
		// check that the returned byte array is not empty
		if len(qrByte) == 0 {
			t.Errorf("NewQR returned empty byte array")
		}

		// test: validate user-provided OTP
		// generate a TOTP object from the input byte slice
		otp, err := twofactor.TOTPFromBytes(encryptedMessage, test.issuer)
		if err != nil {
			t.Fatalf("failed to convert byte slice to TOTP object: %v", err)
		}
		// generate a valid TOTP token for the TOTP object
		validToken, err := otp.OTP()
		if err != nil {
			t.Fatalf("failed to generate valid TOTP token: %v", err)
		}
		// test with a valid token
		_, err = lib.ValidateTOTP(encryptedMessage, test.issuer, validToken)
		if err != nil {
			t.Fatalf("failed to validate a valid TOTP token: %v", err)
		}
		// test with an invalid token
		invalidToken := createInvalidToken(validToken)
		_, err = lib.ValidateTOTP(encryptedMessage, test.issuer, invalidToken)
		if err == nil {
			t.Fatalf("failed to detect an invalid TOTP token: %v", err)
		}
	}
}

func createInvalidToken(in string) (out string) {
	runes := []rune(in) // convert the string to a rune slice

	// reverse the rune slice
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	// convert the rune slice back to a string
	out = string(runes)

	return
}
