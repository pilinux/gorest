package lib

import (
	"crypto"

	"github.com/pilinux/twofactor"
)

// NewTOTP - creates a new TOTP object
// and returns a serialized byte array
func NewTOTP(account string, issuer string, hash crypto.Hash, digits int) ([]byte, error) {
	// TOTP object
	otp, err := twofactor.NewTOTP(account, issuer, hash, digits)
	// internal error
	if err != nil {
		return []byte{}, err
	}

	// serialized byte array
	otpToByte, err := otp.ToBytes()
	// internal error
	if err != nil {
		return []byte{}, err
	}

	return otpToByte, nil
}

// NewQR - creates a byte array containing QR
// code encoded PNG image, with level Q error
// correction, needed for the client apps to
// generate tokens.
func NewQR(encryptedMessage []byte, issuer string) ([]byte, error) {
	otp, err := twofactor.TOTPFromBytes(encryptedMessage, issuer)
	// internal error
	if err != nil {
		return []byte{}, err
	}

	// byte array containing QR code encoded PNG image
	qrBytes, err := otp.QR()
	// internal error
	if err != nil {
		return []byte{}, err
	}

	return qrBytes, nil
}

// ValidateTOTP - validates the user-provided token
func ValidateTOTP(encryptedMessage []byte, issuer string, userInput string) ([]byte, error) {
	otp, err := twofactor.TOTPFromBytes(encryptedMessage, issuer)
	// internal error
	if err != nil {
		return []byte{}, err
	}

	err = otp.Validate(userInput)
	// validation failed
	if err != nil {
		otpToByte, errThis := otp.ToBytes()
		// internal error
		if errThis != nil {
			return []byte{}, errThis
		}

		return otpToByte, err
	}

	// validation successful
	otpToByte, err := otp.ToBytes()
	// internal error
	if err != nil {
		return []byte{}, err
	}

	return otpToByte, nil
}
