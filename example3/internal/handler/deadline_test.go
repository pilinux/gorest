package handler_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pilinux/gorest/example3/internal/handler"
)

// deadlineRecorder is a recorder that also accepts connection deadlines, the
// way a real connection does, and keeps the last of each. gin's ResponseWriter
// unwraps to whatever it was handed, so this is what http.ResponseController
// reaches. A plain httptest recorder has no Unwrap target that accepts
// deadlines, which is why the other tests get ErrNotSupported instead.
type deadlineRecorder struct {
	*httptest.ResponseRecorder
	readDeadline  time.Time
	writeDeadline time.Time
}

// SetReadDeadline records the read deadline the handler asked for.
func (rec *deadlineRecorder) SetReadDeadline(t time.Time) error {
	rec.readDeadline = t
	return nil
}

// SetWriteDeadline records the write deadline the handler asked for.
func (rec *deadlineRecorder) SetWriteDeadline(t time.Time) error {
	rec.writeDeadline = t
	return nil
}

// newDeadlineRecorder returns a recorder that accepts deadlines.
func newDeadlineRecorder() *deadlineRecorder {
	return &deadlineRecorder{ResponseRecorder: httptest.NewRecorder()}
}

// checkDeadline verifies got is FileCryptTimeout from some instant during the
// request. before and after bracket the call, so no tolerance is guessed.
func checkDeadline(t *testing.T, label string, got, before, after time.Time) {
	t.Helper()
	if got.IsZero() {
		t.Fatalf("%s deadline was never set", label)
	}
	earliest := before.Add(handler.FileCryptTimeout)
	latest := after.Add(handler.FileCryptTimeout)
	if got.Before(earliest) || got.After(latest) {
		t.Errorf("%s deadline = %v, want within [%v, %v]", label, got, earliest, latest)
	}
}

// TestHandler_Upload_SetsReadDeadline - both upload shapes bound the body read,
// so a client that stops sending cannot outlive the crypto timeout.
func TestHandler_Upload_SetsReadDeadline(t *testing.T) {
	content := []byte("bounded by a read deadline")

	tests := []struct {
		name    string
		request func(t *testing.T) *http.Request
	}{
		{
			name: "multipart upload",
			request: func(t *testing.T) *http.Request {
				t.Helper()
				body, contentType := multipartBody(t, "note.txt", content)
				req := httptest.NewRequest(http.MethodPost, "/files/encrypt", body)
				req.Header.Set("Content-Type", contentType)
				return req
			},
		},
		{
			name: "raw upload",
			request: func(t *testing.T) *http.Request {
				t.Helper()
				req := httptest.NewRequest(http.MethodPost, "/files/encrypt?name=raw.txt", bytes.NewReader(content))
				req.Header.Set("Content-Type", "application/octet-stream")
				return req
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newFileEngine(t, testUploadLimit)
			rec := newDeadlineRecorder()

			before := time.Now()
			r.ServeHTTP(rec, tc.request(t))
			after := time.Now()

			if rec.Code != http.StatusCreated {
				t.Fatalf("status = %d, want 201 (body: %s)", rec.Code, rec.Body.String())
			}
			checkDeadline(t, "read", rec.readDeadline, before, after)
			if !rec.writeDeadline.IsZero() {
				t.Errorf("write deadline = %v, want none on an upload", rec.writeDeadline)
			}
		})
	}
}

// TestHandler_Download_SetsWriteDeadline - streaming a file back bounds the
// write, so a client that stops reading cannot hold the response open.
func TestHandler_Download_SetsWriteDeadline(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	content := []byte("bounded by a write deadline")

	w := uploadFile(t, r, "stored.txt", content)
	if w.Code != http.StatusCreated {
		t.Fatalf("encrypt status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}

	rec := newDeadlineRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/"+fileIDFromResponse(t, w.Body.Bytes())+"/decrypt", nil)

	before := time.Now()
	r.ServeHTTP(rec, req)
	after := time.Now()

	if rec.Code != http.StatusOK {
		t.Fatalf("decrypt status = %d, want 200 (body: %s)", rec.Code, rec.Body.String())
	}
	if !bytes.Equal(rec.Body.Bytes(), content) {
		t.Error("downloaded content does not match original")
	}
	checkDeadline(t, "write", rec.writeDeadline, before, after)
	if !rec.readDeadline.IsZero() {
		t.Errorf("read deadline = %v, want none on a download", rec.readDeadline)
	}
}
