package handler_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"

	"github.com/pilinux/gorest/example3/internal/handler"
	"github.com/pilinux/gorest/example3/internal/service"
)

// brokenClient is a response writer whose client went away: every body write
// fails.
type brokenClient struct {
	*httptest.ResponseRecorder
}

func (brokenClient) Write([]byte) (int, error) {
	return 0, errors.New("write: broken pipe")
}

// captureLogs records everything logged, down to debug level, until the test
// ends.
func captureLogs(t *testing.T) *logtest.Hook {
	t.Helper()
	std := log.StandardLogger()
	oldHooks := std.ReplaceHooks(make(log.LevelHooks))
	oldLevel, oldOut := std.GetLevel(), std.Out
	t.Cleanup(func() {
		std.ReplaceHooks(oldHooks)
		std.SetLevel(oldLevel)
		std.SetOutput(oldOut)
	})

	hook := new(logtest.Hook)
	std.AddHook(hook)
	std.SetLevel(log.DebugLevel)
	std.SetOutput(io.Discard)
	return hook
}

// TestHandler_Download_ClientGoneIsNotAnError - a client that went away is not
// logged; a damaged file is still an error.
func TestHandler_Download_ClientGoneIsNotAnError(t *testing.T) {
	baseDir := t.TempDir()
	svc := service.NewFileCryptService(newLoadedKeyManager(t), newFakeFileStore(), baseDir, 4<<20)
	api := handler.NewFileCryptAPI(svc)
	r := gin.New()
	r.POST("/files/encrypt/unpadded", api.EncryptFileUnpadded)
	r.GET("/files/:id/decrypt", api.DecryptFile)

	// two chunks, so damage in the second is only found after the body starts
	content := bytes.Repeat([]byte("x"), 1<<20+100)
	upload := func(t *testing.T) string {
		t.Helper()
		w := uploadMultipartTo(t, r, "/files/encrypt/unpadded", "log.txt", content)
		if w.Code != http.StatusCreated {
			t.Fatalf("encrypt status = %d, want 201 (body: %s)", w.Code, w.Body.String())
		}
		return fileIDFromResponse(t, w.Body.Bytes())
	}

	t.Run("client gone", func(t *testing.T) {
		id := upload(t)
		hook := captureLogs(t)

		rec := brokenClient{httptest.NewRecorder()}
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/files/"+id+"/decrypt", nil))

		if entries := hook.AllEntries(); len(entries) != 0 {
			t.Errorf("logged %q, want nothing for a client that went away", entries[0].Message)
		}
	})

	t.Run("damaged file", func(t *testing.T) {
		id := upload(t)
		path := filepath.Join(baseDir, id+".enc")
		blob, err := os.ReadFile(path) // #nosec G304 -- test-owned temp path
		if err != nil {
			t.Fatalf("read error: %v", err)
		}
		blob[len(blob)-1] ^= 0x01
		if err := os.WriteFile(path, blob, 0o600); err != nil {
			t.Fatalf("write error: %v", err)
		}
		hook := captureLogs(t)

		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/files/"+id+"/decrypt", nil))

		entry := hook.LastEntry()
		if entry == nil || entry.Message != "DecryptFile.h.1" || entry.Level != log.ErrorLevel {
			t.Errorf("last log = %v, want DecryptFile.h.1 at error level", entry)
		}
	})
}
