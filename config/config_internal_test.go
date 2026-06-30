package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeConfigDir(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		wantErr error
	}{
		{
			name: "valid relative dir",
			dir:  "tmp/config-test",
		},
		{
			name:    "path traversal rejected",
			dir:     "../tmp/config-test",
			wantErr: errInvalidConfigDir,
		},
		{
			name:    "absolute dir rejected",
			dir:     filepath.Join(string(filepath.Separator), "tmp", "config-test"),
			wantErr: errInvalidConfigDir,
		},
		{
			name: "parent segments resolved inside workspace",
			dir:  "tmp/../tmp/config-test",
		},
		{
			name: "current dir allowed",
			dir:  ".",
		},
		{
			name:    "multiple parent segments rejected",
			dir:     "tmp/../../tmp/config-test",
			wantErr: errInvalidConfigDir,
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			path, err := sanitizeConfigDir(tc.dir)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error: %v, got: %v", tc.wantErr, err)
				}
				if path != "" {
					t.Fatalf("expected empty path, got %q", path)
				}
				return
			}

			if err != nil {
				t.Fatalf("sanitizeConfigDir() error = %v", err)
			}

			cwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("os.Getwd() error = %v", err)
			}

			want := filepath.Join(cwd, filepath.Clean(tc.dir))
			if path != want {
				t.Fatalf("expected %q, got %q", want, path)
			}
		})
	}
}

func TestEnsureConfigDir(t *testing.T) {
	tempRoot := t.TempDir()

	prevWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd() error = %v", err)
	}

	if err = os.Chdir(tempRoot); err != nil {
		t.Fatalf("os.Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(prevWD); chdirErr != nil {
			t.Fatalf("restore working directory error = %v", chdirErr)
		}
	})

	if err = ensureConfigDir(filepath.Join("tmp", "nested")); err != nil {
		t.Fatalf("ensureConfigDir() error = %v", err)
	}

	createdPath := filepath.Join(tempRoot, "tmp", "nested")
	info, err := os.Stat(createdPath)
	if err != nil {
		t.Fatalf("os.Stat() error = %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected %q to be a directory", createdPath)
	}

	// call again on existing directory — should return nil immediately (stat succeeds)
	if err = ensureConfigDir(filepath.Join("tmp", "nested")); err != nil {
		t.Fatalf("ensureConfigDir() on existing dir error = %v", err)
	}

	if err = ensureConfigDir(filepath.Join("..", "escape")); !errors.Is(err, errInvalidConfigDir) {
		t.Fatalf("expected errInvalidConfigDir for traversal path, got %v", err)
	}
}

func TestCanonicalJWTAlg(t *testing.T) {
	tests := []struct {
		name    string
		alg     string
		want    string
		wantErr bool
	}{
		// every supported value in the switch is covered
		{name: "hs256", alg: "hs256", want: "HS256"},
		{name: "hs384", alg: "hs384", want: "HS384"},
		{name: "hs512", alg: "hs512", want: "HS512"},
		{name: "es256", alg: "es256", want: "ES256"},
		{name: "es384", alg: "es384", want: "ES384"},
		{name: "es512", alg: "es512", want: "ES512"},
		{name: "eddsa", alg: "eddsa", want: "EdDSA"},
		{name: "rs256", alg: "rs256", want: "RS256"},
		{name: "rs384", alg: "rs384", want: "RS384"},
		{name: "rs512", alg: "rs512", want: "RS512"},
		// matched case-insensitively: upper and mixed case map to canonical
		{name: "uppercase HS256", alg: "HS256", want: "HS256"},
		{name: "canonical EdDSA", alg: "EdDSA", want: "EdDSA"},
		{name: "uppercase EDDSA", alg: "EDDSA", want: "EdDSA"},
		{name: "mixed case Rs512", alg: "Rs512", want: "RS512"},
		// default branch: unsupported and empty input
		{name: "unsupported value", alg: "any", wantErr: true},
		{name: "empty value", alg: "", wantErr: true},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(tc.name, func(t *testing.T) {
			got, err := canonicalJWTAlg(tc.alg)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("canonicalJWTAlg(%q): expected error, got nil", tc.alg)
				}
				if got != "" {
					t.Fatalf("canonicalJWTAlg(%q): expected empty result on error, got %q", tc.alg, got)
				}
				return
			}

			if err != nil {
				t.Fatalf("canonicalJWTAlg(%q): unexpected error: %v", tc.alg, err)
			}
			if got != tc.want {
				t.Fatalf("canonicalJWTAlg(%q) = %q, want %q", tc.alg, got, tc.want)
			}
		})
	}
}
