package lib_test

import (
	"testing"

	"github.com/pilinux/gorest/lib"
)

type hashPassTest struct {
	name         string
	config       lib.HashPassConfig
	password     string
	secret       string
	expectedErr  bool
	nonExpectedH string
}

func TestHashPass(t *testing.T) {
	tests := []hashPassTest{
		{
			name: "blank password, no secret",
			config: lib.HashPassConfig{
				Memory:      32,
				Iterations:  2,
				Parallelism: 1,
				SaltLength:  16,
				KeyLength:   16,
			},
			password:     "",
			expectedErr:  false,
			nonExpectedH: "", // empty string
		},
		{
			name: "with password, no secret",
			config: lib.HashPassConfig{
				Memory:      32,
				Iterations:  2,
				Parallelism: 1,
				SaltLength:  16,
				KeyLength:   16,
			},
			password:     "password123",
			expectedErr:  false,
			nonExpectedH: "", // empty string
		},
		{
			name: "with password, with secret",
			config: lib.HashPassConfig{
				Memory:      32,
				Iterations:  2,
				Parallelism: 1,
				SaltLength:  16,
				KeyLength:   16,
			},
			password:     "password123",
			secret:       "secret123",
			expectedErr:  false,
			nonExpectedH: "", // empty string
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h, err := lib.HashPass(test.config, test.password, test.secret)
			if (err != nil) != test.expectedErr {
				t.Errorf("unexpected error: got %v, want %v", err, test.expectedErr)
			}
			if h == test.nonExpectedH {
				t.Errorf("unexpected hash: got %v, want %v", h, test.nonExpectedH)
			}
		})
	}
}
