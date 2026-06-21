package database_test

import (
	"testing"

	gdb "github.com/pilinux/gorest/database"
)

// TestQuotePostgresDSNValue verifies that values are quoted and that
// single quotes and backslashes are escaped per the libpq keyword/value
// DSN format.
func TestQuotePostgresDSNValue(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "empty value",
			in:   "",
			want: `''`,
		},
		{
			name: "simple value",
			in:   "mydb",
			want: `'mydb'`,
		},
		{
			name: "value with space",
			in:   "pa ss",
			want: `'pa ss'`,
		},
		{
			name: "value with single quote",
			in:   "pa'ss",
			want: `'pa\'ss'`,
		},
		{
			name: "value with backslash",
			in:   `pa\ss`,
			want: `'pa\\ss'`,
		},
		{
			name: "value with backslash and quote",
			in:   `p\a'ss`,
			want: `'p\\a\'ss'`,
		},
		{
			name: "value with backslash and quote one after another",
			in:   `p\'a'ss`,
			want: `'p\\\'a\'ss'`,
		},
		{
			name: "value with equals and spaces",
			in:   "name = value",
			want: `'name = value'`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := gdb.QuotePostgresDSNValue(tc.in)
			if got != tc.want {
				t.Errorf("QuotePostgresDSNValue(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
