package lib

import (
	"crypto/rand"
	"math/big"
)

// SecureRandomNumber generates a cryptographically secure pseudo-random number.
// To generate a random number consisting of x number of digits, pass x as
// the parameter. For example, SecureRandomNumber(3) will generate a number
// between 100 and 999.
func SecureRandomNumber(totalDigit uint64) uint64 {
	if totalDigit == 0 {
		return 0
	}

	// Compute the bounds with big.Int.Exp for exact powers of ten. Using
	// math.Pow + int64 loses precision for totalDigit >= 16, since float64
	// only has 53-bit integer precision, and overflows int64 for
	// totalDigit >= 19. Both yield invalid bounds (a garbage or non-positive
	// maxVal), which can make rand.Int panic.
	ten := big.NewInt(10)
	// minVal = 10^(totalDigit-1)
	minVal := new(big.Int).Exp(ten, new(big.Int).SetUint64(totalDigit-1), nil)
	// maxVal = 10^totalDigit - 1
	maxVal := new(big.Int).Sub(
		new(big.Int).Exp(ten, new(big.Int).SetUint64(totalDigit), nil),
		big.NewInt(1),
	)

	// Draw uniformly from the inclusive range [minVal, maxVal]. rand.Int yields
	// [0, span), so sample a span of size (maxVal-minVal+1) and shift by minVal.
	// This covers both endpoints (e.g. 100 and 999 for totalDigit==3) and avoids
	// the rejection-sampling loop.
	span := new(big.Int).Add(new(big.Int).Sub(maxVal, minVal), big.NewInt(1))

	var result *big.Int
	for {
		x, err := rand.Int(rand.Reader, span)
		if err == nil {
			result = new(big.Int).Add(x, minVal)
			break
		}
	}

	return result.Uint64()
}
