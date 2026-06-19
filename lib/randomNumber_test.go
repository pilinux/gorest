package lib_test

import (
	"strconv"
	"testing"

	"github.com/pilinux/gorest/lib"
)

func TestSecureRandomNumber(t *testing.T) {
	tests := []struct {
		name       string
		totalDigit uint64
		minimum    uint64
		maximum    uint64
	}{
		{
			name:       "test case 1",
			totalDigit: 3,
			minimum:    100,
			maximum:    999,
		},
		{
			name:       "test case 2",
			totalDigit: 5,
			minimum:    10000,
			maximum:    99999,
		},
		{
			name:       "test case 3",
			totalDigit: 1,
			minimum:    0,
			maximum:    9,
		},
		{
			name:       "test case 4",
			totalDigit: 0,
			minimum:    0,
			maximum:    0,
		},
		{
			// smallest multi-digit range
			name:       "edge: two digits",
			totalDigit: 2,
			minimum:    10,
			maximum:    99,
		},
		{
			// float64 (53-bit mantissa) can no longer represent 10^16 exactly
			name:       "edge: float64 precision boundary",
			totalDigit: 16,
			minimum:    1000000000000000,
			maximum:    9999999999999999,
		},
		{
			// 10^19 overflows int64
			name:       "edge: int64 overflow boundary",
			totalDigit: 19,
			minimum:    1000000000000000000,
			maximum:    9999999999999999999,
		},
	}

	for i := range tests {
		tt := tests[i]
		t.Run(tt.name, func(t *testing.T) {
			got := lib.SecureRandomNumber(tt.totalDigit)

			if got < tt.minimum || got > tt.maximum {
				t.Errorf("SecureRandomNumber() = %v, want in between %v and %v", got, tt.minimum, tt.maximum)
			}
		})
	}
}

// TestSecureRandomNumberEndpoints verifies that both endpoints of the
// documented inclusive range are reachable.
func TestSecureRandomNumberEndpoints(t *testing.T) {
	const (
		totalDigit = 3
		minVal     = 100
		maxVal     = 999
		iterations = 100000
	)

	hitMin, hitMax := false, false
	for range iterations {
		got := lib.SecureRandomNumber(totalDigit)
		if got < minVal || got > maxVal {
			t.Fatalf("SecureRandomNumber(%d) = %v, out of range [%d, %d]", totalDigit, got, minVal, maxVal)
		}
		if got == minVal {
			hitMin = true
		}
		if got == maxVal {
			hitMax = true
		}
		if hitMin && hitMax {
			break
		}
	}

	if !hitMin {
		t.Errorf("minimum endpoint %d never produced over %d draws", minVal, iterations)
	}
	if !hitMax {
		t.Errorf("maximum endpoint %d never produced over %d draws", maxVal, iterations)
	}
}

// TestSecureRandomNumberDigitCount verifies the result always has exactly the
// requested number of digits (no shorter/longer values leak through).
func TestSecureRandomNumberDigitCount(t *testing.T) {
	for _, totalDigit := range []uint64{1, 2, 3, 5, 9} {
		for range 1000 {
			got := lib.SecureRandomNumber(totalDigit)
			gotDigits := uint64(len(strconv.FormatUint(got, 10)))
			if gotDigits != totalDigit {
				t.Fatalf("SecureRandomNumber(%d) = %v has %d digits, want %d", totalDigit, got, gotDigits, totalDigit)
			}
		}
	}
}
