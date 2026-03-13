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

	if err = ensureConfigDir(filepath.Join("..", "escape")); !errors.Is(err, errInvalidConfigDir) {
		t.Fatalf("expected errInvalidConfigDir for traversal path, got %v", err)
	}
}
