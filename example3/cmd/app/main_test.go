package main

import (
	"math"
	"strconv"
	"testing"
)

// maxUploadSizeMiB is the largest MAX_UPLOAD_SIZE_MB that still fits in an
// int64 once shifted to bytes.
const maxUploadSizeMiB = math.MaxInt64 >> 20

// TestMaxUploadSize - MAX_UPLOAD_SIZE_MB is converted to bytes, and anything
// that would overflow the shift falls back to the default.
func TestMaxUploadSize(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want int64
	}{
		{
			name: "unset falls back to the default",
			env:  "",
			want: defaultMaxUploadSize,
		},
		{
			name: "a plain value is converted to bytes",
			env:  "10",
			want: 10 << 20,
		},
		{
			name: "surrounding spaces are trimmed",
			env:  "  10  ",
			want: 10 << 20,
		},
		{
			name: "the largest non-overflowing value is accepted",
			env:  strconv.FormatInt(maxUploadSizeMiB, 10),
			want: maxUploadSizeMiB << 20,
		},
		{
			name: "one MiB past the limit overflows, so the default is used",
			env:  strconv.FormatInt(maxUploadSizeMiB+1, 10),
			want: defaultMaxUploadSize,
		},
		{
			name: "math.MaxInt64 overflows, so the default is used",
			env:  strconv.FormatInt(math.MaxInt64, 10),
			want: defaultMaxUploadSize,
		},
		{
			name: "zero falls back to the default",
			env:  "0",
			want: defaultMaxUploadSize,
		},
		{
			name: "a negative value falls back to the default",
			env:  "-5",
			want: defaultMaxUploadSize,
		},
		{
			name: "an unparsable value falls back to the default",
			env:  "abc",
			want: defaultMaxUploadSize,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("MAX_UPLOAD_SIZE_MB", tc.env)

			got := maxUploadSize()
			if got != tc.want {
				t.Errorf("maxUploadSize() = %d, want %d", got, tc.want)
			}
			if got <= 0 {
				t.Errorf("maxUploadSize() = %d, want a positive limit", got)
			}
		})
	}
}
