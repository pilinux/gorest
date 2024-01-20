package lib_test

import (
	"testing"

	"github.com/pilinux/gorest/lib"
)

// TestInitRateLimiter - test the InitRateLimiter function
func TestInitRateLimiter(t *testing.T) {
	tests := []struct {
		name            string
		rateLimit       string
		trustedProxy    string
		expectedLimiter bool
		expectedErr     bool
	}{
		{
			name:            "Empty rate limit, should pass",
			rateLimit:       "",
			trustedProxy:    "",
			expectedLimiter: false,
			expectedErr:     false,
		},
		{
			name:            "Invalid rate limit, should fail",
			rateLimit:       "10",
			trustedProxy:    "",
			expectedLimiter: false,
			expectedErr:     true,
		},
		{
			name:            "Valid rate limit, should pass",
			rateLimit:       "10-S",
			trustedProxy:    "",
			expectedLimiter: true,
			expectedErr:     false,
		},
		{
			name:            "Valid rate limit with trusted proxy, should pass",
			rateLimit:       "10-S",
			trustedProxy:    "X-Real-Ip",
			expectedLimiter: true,
			expectedErr:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			limiterInstance, err := lib.InitRateLimiter(test.rateLimit, test.trustedProxy)

			if (err != nil) != test.expectedErr {
				t.Errorf("expected error: %v, got: %v", test.expectedErr, err)
			}
			if (err == nil) == test.expectedErr {
				t.Errorf("expected error: %v, got: %v", test.expectedErr, err)
			}

			if (limiterInstance != nil) != test.expectedLimiter {
				t.Errorf("expected limiterInstance: %v, got: %v", test.expectedLimiter, limiterInstance)
			}
			if (limiterInstance == nil) == test.expectedLimiter {
				t.Errorf("expected limiterInstance: %v, got: %v", test.expectedLimiter, limiterInstance)
			}
		})
	}
}
