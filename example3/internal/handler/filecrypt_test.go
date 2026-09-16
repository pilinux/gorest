package handler_test

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/pilinux/gorest/example3/internal/handler"
	"github.com/pilinux/gorest/example3/internal/service"
)

// testUploadLimit is the upload cap used by the handler tests.
const testUploadLimit = 1 << 20 // 1 MiB

// newFileEngine builds a gin engine with the file crypto routes, backed by an
// in-memory metadata store and a throwaway base directory.
func newFileEngine(t *testing.T, maxUploadSize int64) *gin.Engine {
	t.Helper()
	km := newLoadedKeyManager(t)
	fileSrv := service.NewFileCryptService(km, newFakeFileStore(), t.TempDir(), maxUploadSize)
	api := handler.NewFileCryptAPI(fileSrv)

	r := gin.New()
	r.POST("/files/encrypt", api.EncryptFile)
	r.POST("/files/encrypt/unpadded", api.EncryptFileUnpadded)
	r.GET("/files/:id/decrypt", api.DecryptFile)
	r.DELETE("/files/:id", api.DeleteFile)
	return r
}

// multipartBody builds a multipart body holding an ordinary form field followed
// by the file, so the handler's part scan is exercised as well.
func multipartBody(t *testing.T, filename string, content []byte) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	if err := mw.WriteField("note", "a field before the file"); err != nil {
		t.Fatalf("WriteField error: %v", err)
	}
	fw, err := mw.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile error: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	return &body, mw.FormDataContentType()
}

// uploadFile posts a multipart file upload and returns the recorder.
func uploadFile(t *testing.T, r *gin.Engine, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	body, contentType := multipartBody(t, filename, content)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt", body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)
	return w
}

// fileIDFromResponse extracts message.fileId from an encrypt response.
func fileIDFromResponse(t *testing.T, body []byte) string {
	t.Helper()
	var parsed struct {
		Message struct {
			FileID string `json:"fileId"`
		} `json:"message"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("unmarshal error: %v (body: %s)", err, body)
	}
	return parsed.Message.FileID
}

func TestHandler_FileRoundTrip(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	content := []byte("confidential payload 1234567890")

	// encrypt/upload
	w := uploadFile(t, r, "report.txt", content)
	if w.Code != http.StatusCreated {
		t.Fatalf("encrypt status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}
	fileID := fileIDFromResponse(t, w.Body.Bytes())
	if fileID == "" {
		t.Fatal("empty file id")
	}

	// decrypt/download
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/"+fileID+"/decrypt", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("decrypt status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if !bytes.Equal(w.Body.Bytes(), content) {
		t.Error("downloaded content does not match original")
	}
	if cd := w.Header().Get("Content-Disposition"); cd != `attachment; filename="report.txt"` {
		t.Errorf("Content-Disposition = %q", cd)
	}
	if cl := w.Header().Get("Content-Length"); cl != strconv.Itoa(len(content)) {
		t.Errorf("Content-Length = %q, want %d", cl, len(content))
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Errorf("Content-Type = %q", ct)
	}

	// delete
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/files/"+fileID, nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want 200", w.Code)
	}

	// second decrypt must now be 404
	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/files/"+fileID+"/decrypt", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("decrypt after delete status = %d, want 404", w.Code)
	}
}

func TestHandler_FileRoundTrip_LargerThanOneChunk(t *testing.T) {
	r := newFileEngine(t, 4<<20)
	// more than the 1 MiB chunk size, so the upload is sealed and served back
	// as several chunks
	content := make([]byte, (1<<20)+512)
	if _, err := rand.Read(content); err != nil {
		t.Fatalf("rand.Read error: %v", err)
	}

	w := uploadFile(t, r, "large.bin", content)
	if w.Code != http.StatusCreated {
		t.Fatalf("encrypt status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}
	fileID := fileIDFromResponse(t, w.Body.Bytes())

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/"+fileID+"/decrypt", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("decrypt status = %d, want 200", w.Code)
	}
	if !bytes.Equal(w.Body.Bytes(), content) {
		t.Error("downloaded content does not match original")
	}
}

func TestHandler_FileEncrypt_TooLarge(t *testing.T) {
	const limit = 1 << 12
	r := newFileEngine(t, limit)

	w := uploadFile(t, r, "toobig.bin", bytes.Repeat([]byte("x"), limit+1))
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413 (body: %s)", w.Code, w.Body.String())
	}
}

// TestHandler_FileEncrypt_BodyTooLarge covers the cap on the whole request
// body: every part here is an ordinary form field well inside the file limit,
// so only the body cap can stop them. Without it the handler would read all of
// them off the wire while hunting for a file part that never arrives.
func TestHandler_FileEncrypt_BodyTooLarge(t *testing.T) {
	const limit = 1 << 12
	r := newFileEngine(t, limit)

	// the body may weigh limit + 1 MiB, so pad past that with junk fields
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	junk := bytes.Repeat([]byte("x"), 256<<10)
	for i := range 6 {
		if err := mw.WriteField("junk"+strconv.Itoa(i), string(junk)); err != nil {
			t.Fatalf("WriteField error: %v", err)
		}
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413 (body: %s)", w.Code, w.Body.String())
	}
}

// TestHandler_FileEncrypt_BodyTooLargeDuringFile is the same cap seen from the
// other side: the file is comfortably inside its own limit, but the fields sent
// ahead of it have already spent most of the body budget, so the cap runs out
// while the file is being read. That failure surfaces from the seal rather than
// from the part scan, and must still be reported as 413.
func TestHandler_FileEncrypt_BodyTooLargeDuringFile(t *testing.T) {
	const limit = 1 << 12
	r := newFileEngine(t, limit)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	junk := bytes.Repeat([]byte("x"), 256<<10)
	for i := range 4 { // ~1 MiB of fields: just under the cap, not past it
		if err := mw.WriteField("junk"+strconv.Itoa(i), string(junk)); err != nil {
			t.Fatalf("WriteField error: %v", err)
		}
	}
	fw, err := mw.CreateFormFile("file", "small.bin")
	if err != nil {
		t.Fatalf("CreateFormFile error: %v", err)
	}
	// well within the 4 KiB file limit, yet past what is left of the body cap
	if _, err := fw.Write(bytes.Repeat([]byte("y"), limit-96)); err != nil {
		t.Fatalf("write error: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413 (body: %s)", w.Code, w.Body.String())
	}
}

func TestHandler_FileEncrypt_NoFile(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt", nil)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=none")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandler_FileEncrypt_NotMultipart(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt", bytes.NewReader([]byte(`{"plaintext":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want 415", w.Code)
	}
}

func TestHandler_FileEncrypt_FieldsOnly(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := mw.WriteField("note", "no file here"); err != nil {
		t.Fatalf("WriteField error: %v", err)
	}
	_ = mw.Close()

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestHandler_FileEncrypt_EmptyFile(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)

	w := uploadFile(t, r, "empty.txt", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestHandler_FileDecrypt_InvalidID(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	w := httptest.NewRecorder()
	// "not-hex" fails the file-id format check -> 400
	req := httptest.NewRequest(http.MethodGet, "/files/not-hex/decrypt", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// uploadRaw posts the file as the whole request body, the way a client that
// knows its length does. httptest sets Content-Length from a *bytes.Reader.
func uploadRaw(t *testing.T, r *gin.Engine, target string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, target, bytes.NewReader(content))
	req.Header.Set("Content-Type", "application/octet-stream")
	if req.ContentLength != int64(len(content)) {
		t.Fatalf("ContentLength = %d, want %d", req.ContentLength, len(content))
	}
	r.ServeHTTP(w, req)
	return w
}

// downloadFile fetches a stored file.
func downloadFile(t *testing.T, r *gin.Engine, fileID string) *httptest.ResponseRecorder {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/"+fileID+"/decrypt", nil)
	r.ServeHTTP(w, req)
	return w
}

func TestHandler_RawUpload_RoundTrip(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	content := []byte("sent as the whole body, length declared up front")

	w := uploadRaw(t, r, "/files/encrypt?name=raw.txt", content)
	if w.Code != http.StatusCreated {
		t.Fatalf("encrypt status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}

	w = downloadFile(t, r, fileIDFromResponse(t, w.Body.Bytes()))
	if w.Code != http.StatusOK {
		t.Fatalf("decrypt status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	if !bytes.Equal(w.Body.Bytes(), content) {
		t.Error("downloaded content does not match original")
	}
	if cd := w.Header().Get("Content-Disposition"); cd != `attachment; filename="raw.txt"` {
		t.Errorf("Content-Disposition = %q", cd)
	}
}

// TestHandler_RawUpload_NameFromContentDisposition covers the other way to name
// a raw upload.
func TestHandler_RawUpload_NameFromContentDisposition(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt", bytes.NewReader([]byte("named by header")))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Disposition", `attachment; filename="report.pdf"`)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}

	w = downloadFile(t, r, fileIDFromResponse(t, w.Body.Bytes()))
	if cd := w.Header().Get("Content-Disposition"); cd != `attachment; filename="report.pdf"` {
		t.Errorf("Content-Disposition = %q", cd)
	}
}

// TestHandler_RawUpload_TooLarge: a declared length past the limit is refused
// before the body is read.
func TestHandler_RawUpload_TooLarge(t *testing.T) {
	const limit = 1 << 12
	r := newFileEngine(t, limit)

	w := uploadRaw(t, r, "/files/encrypt", bytes.Repeat([]byte("x"), limit+1))
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413 (body: %s)", w.Code, w.Body.String())
	}
	// "file is too large", not "request body is too large": it was the declared
	// length that stopped it, not the cap running out mid-read
	if !strings.Contains(w.Body.String(), "file is too large") {
		t.Errorf("body = %s, want a file-too-large message", w.Body.String())
	}
}

// TestHandler_RawUpload_Chunked is a raw body with no length to declare. It
// falls back to the same measured path a multipart part takes.
func TestHandler_RawUpload_Chunked(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	content := []byte("streamed with no Content-Length")

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt?name=chunked.bin", bytes.NewReader(content))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = -1 // what a chunked body looks like to the server
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}

	w = downloadFile(t, r, fileIDFromResponse(t, w.Body.Bytes()))
	if !bytes.Equal(w.Body.Bytes(), content) {
		t.Error("downloaded content does not match original")
	}
}

// TestHandler_RawUpload_LyingLength covers a client whose body is shorter than
// the Content-Length it declared: the padding was sized for the promise, so the
// upload is rejected rather than stored short.
func TestHandler_RawUpload_LyingLength(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt", bytes.NewReader([]byte("only this much")))
	req.Header.Set("Content-Type", "application/octet-stream")
	req.ContentLength = 1000
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestHandler_RawUpload_Empty(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)

	w := uploadRaw(t, r, "/files/encrypt", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

// TestHandler_RawUpload_ExactlyAtLimit: the body cap allows the limit itself,
// including the one-byte probe the sealer makes past the declared length.
func TestHandler_RawUpload_ExactlyAtLimit(t *testing.T) {
	const limit = 1 << 12
	r := newFileEngine(t, limit)

	w := uploadRaw(t, r, "/files/encrypt", bytes.Repeat([]byte("x"), limit))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}

	w = downloadFile(t, r, fileIDFromResponse(t, w.Body.Bytes()))
	if got := len(w.Body.Bytes()); got != limit {
		t.Errorf("downloaded %d bytes, want %d", got, limit)
	}
}

// uploadMultipartTo posts a multipart file upload to an arbitrary route.
func uploadMultipartTo(t *testing.T, r *gin.Engine, target, filename string, content []byte) *httptest.ResponseRecorder {
	t.Helper()
	body, contentType := multipartBody(t, filename, content)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, target, body)
	req.Header.Set("Content-Type", contentType)
	r.ServeHTTP(w, req)
	return w
}

// TestHandler_UnpaddedRoundTrip covers both body shapes on the unpadded route,
// and the shared decrypt route serving them back.
func TestHandler_UnpaddedRoundTrip(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	content := []byte("no padding, sealed as it arrives")

	tests := []struct {
		name string
		post func() *httptest.ResponseRecorder
	}{
		{
			name: "multipart",
			post: func() *httptest.ResponseRecorder {
				return uploadMultipartTo(t, r, "/files/encrypt/unpadded", "form.txt", content)
			},
		},
		{
			name: "raw",
			post: func() *httptest.ResponseRecorder {
				return uploadRaw(t, r, "/files/encrypt/unpadded?name=raw.txt", content)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := tc.post()
			if w.Code != http.StatusCreated {
				t.Fatalf("encrypt status = %d, want 201 (body: %s)", w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), `"unpadded":true`) {
				t.Errorf("response does not report the file as unpadded: %s", w.Body.String())
			}

			w = downloadFile(t, r, fileIDFromResponse(t, w.Body.Bytes()))
			if w.Code != http.StatusOK {
				t.Fatalf("decrypt status = %d, want 200 (body: %s)", w.Code, w.Body.String())
			}
			if !bytes.Equal(w.Body.Bytes(), content) {
				t.Error("downloaded content does not match original")
			}
			if cl := w.Header().Get("Content-Length"); cl != strconv.Itoa(len(content)) {
				t.Errorf("Content-Length = %q, want %d", cl, len(content))
			}
		})
	}
}

// TestHandler_PaddedResponseHasNoFlag: the padded route is the default, so its
// records carry no format field at all.
func TestHandler_PaddedResponseHasNoFlag(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)

	w := uploadFile(t, r, "padded.txt", []byte("padded by default"))
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (body: %s)", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "unpadded") {
		t.Errorf("padded response mentions the format flag: %s", w.Body.String())
	}
}

// TestHandler_UnpaddedBothFormatsCoexist: two files, one of each format, served
// back by the same decrypt route.
func TestHandler_UnpaddedBothFormatsCoexist(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)
	padded := []byte("this one is padded")
	plain := []byte("this one is not")

	w := uploadFile(t, r, "padded.txt", padded)
	paddedID := fileIDFromResponse(t, w.Body.Bytes())
	w = uploadMultipartTo(t, r, "/files/encrypt/unpadded", "plain.txt", plain)
	plainID := fileIDFromResponse(t, w.Body.Bytes())

	if got := downloadFile(t, r, paddedID).Body.Bytes(); !bytes.Equal(got, padded) {
		t.Error("padded file came back wrong")
	}
	if got := downloadFile(t, r, plainID).Body.Bytes(); !bytes.Equal(got, plain) {
		t.Error("unpadded file came back wrong")
	}
}

func TestHandler_UnpaddedEncrypt_Empty(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)

	w := uploadMultipartTo(t, r, "/files/encrypt/unpadded", "empty.txt", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (body: %s)", w.Code, w.Body.String())
	}
}

func TestHandler_UnpaddedEncrypt_TooLarge(t *testing.T) {
	const limit = 1 << 12
	r := newFileEngine(t, limit)

	w := uploadMultipartTo(t, r, "/files/encrypt/unpadded", "toobig.bin", bytes.Repeat([]byte("x"), limit+1))
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413 (body: %s)", w.Code, w.Body.String())
	}
}

func TestHandler_UnpaddedEncrypt_NotAnUpload(t *testing.T) {
	r := newFileEngine(t, testUploadLimit)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/files/encrypt/unpadded", bytes.NewReader([]byte(`{"plaintext":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnsupportedMediaType {
		t.Errorf("status = %d, want 415", w.Code)
	}
}
