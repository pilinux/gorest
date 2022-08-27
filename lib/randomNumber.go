package lib

import (
	"crypto/rand"
	"math"
	"math/big"
)

// SecureRandomNumber generates cryptographically secure pseudo-random number.
// To generate a random number consisting of x number of digits, pass x as
// the parameter. For example, SecureRandomNumber(3) will generate a number
// between 100 and 999
func SecureRandomNumber(totalDigit uint64) uint64 {
	if totalDigit == 0 {
		return 0
	}

	var result *big.Int

	min := big.NewInt(int64(math.Pow(10, float64(totalDigit)-1)))
	max := big.NewInt(int64(math.Pow(10, float64(totalDigit)) - 1))

	for {
		x, err := rand.Int(rand.Reader, max)

		if err == nil {
			if x.Cmp(min) == +1 {
				result = x
				break
			}
		}
	}

	return result.Uint64()
}
