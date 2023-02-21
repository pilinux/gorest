package lib_test

import (
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lib.SecureRandomNumber(tt.totalDigit)

			if got < tt.minimum || got > tt.maximum {
				t.Errorf("SecureRandomNumber() = %v, want in between %v and %v", got, tt.minimum, tt.maximum)
			}
		})
	}
}
