package service

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pilinux/crypt/envelope"
)

func TestValidFileID(t *testing.T) {
	valid := []string{
		"0123456789abcdef",                 // 16
		"0123456789abcdef0123456789abcdef", // 32
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", // 64
	}
	for _, id := range valid {
		if !validFileID(id) {
			t.Errorf("validFileID(%q) = false, want true", id)
		}
	}

	invalid := []string{
		"",                                 // empty
		"short",                            // too short / non-hex
		"0123456789ABCDEF",                 // uppercase not allowed
		"../../etc/passwd",                 // path traversal
		"0123456789abcdef0123456789abcdeg", // 'g' is not hex
		"0123456789abcdef/",                // separator
	}
	for _, id := range invalid {
		if validFileID(id) {
			t.Errorf("validFileID(%q) = true, want false", id)
		}
	}
}

func TestEncryptedFilePath(t *testing.T) {
	t.Run("validID", func(t *testing.T) {
		id := "0123456789abcdef0123456789abcdef"
		got, err := encryptedFilePath("/base", id)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		want := filepath.Join("/base", id+encryptedFileExt)
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("invalidIDRejected", func(t *testing.T) {
		if _, err := encryptedFilePath("/base", "../evil"); !errors.Is(err, ErrInvalidFileID) {
			t.Errorf("err = %v, want ErrInvalidFileID", err)
		}
	})
}

func TestEnsureDir(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, "a", "b", "c")

	if err := ensureDir(target); err != nil {
		t.Fatalf("ensureDir error: %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if !info.IsDir() {
		t.Error("ensureDir did not create a directory")
	}

	// calling again on an existing directory must be a no-op, not an error
	if err := ensureDir(target); err != nil {
		t.Errorf("ensureDir on existing dir error: %v", err)
	}
}

// eofWithDataReader hands out all of its content in a single read that also
// reports io.EOF, the way a multipart part does at the end of an upload.
type eofWithDataReader struct {
	data []byte
}

func (r *eofWithDataReader) Read(p []byte) (int, error) {
	n := copy(p, r.data)
	r.data = r.data[n:]
	if len(r.data) == 0 {
		return n, io.EOF
	}
	return n, nil
}

func TestLimitedReader(t *testing.T) {
	const limit = 8
	source := strings.Repeat("x", 32)

	tests := []struct {
		name string
		size int
		want error
	}{
		{name: "underLimit", size: limit - 1},
		{name: "atLimit", size: limit},
		{name: "overLimit", size: limit + 1, want: ErrFileTooLarge},
		{name: "farOverLimit", size: 4 * limit, want: ErrFileTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newLimitedReader(strings.NewReader(source[:tt.size]), limit)
			got, err := io.ReadAll(r)
			if !errors.Is(err, tt.want) {
				t.Fatalf("err = %v, want %v", err, tt.want)
			}
			// nothing is truncated silently: what comes out is either the whole
			// input or an error
			if tt.want == nil && len(got) != tt.size {
				t.Errorf("read %d bytes, want %d", len(got), tt.size)
			}
		})
	}

	// a source that reports the end of its data in the same read as the data
	// itself must not slip past the limit
	t.Run("dataWithEOF", func(t *testing.T) {
		r := newLimitedReader(&eofWithDataReader{data: []byte(source[:limit+1])}, limit)
		if _, err := io.ReadAll(r); !errors.Is(err, ErrFileTooLarge) {
			t.Errorf("err = %v, want ErrFileTooLarge", err)
		}
	})
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "report.pdf", want: "report.pdf"},
		{name: "stripsPath", in: "/etc/passwd", want: "passwd"},
		{name: "stripsWindowsPath", in: `C:\secret\file.txt`, want: "file.txt"},
		{name: "empty", in: "", want: "file"},
		{name: "dotsOnly", in: "..", want: "file"},
		{name: "removesQuotesAndControls", in: "a\"b\nc.txt", want: "abc.txt"},
		{name: "trimsSpace", in: "  spaced.txt  ", want: "spaced.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeName(tt.in); got != tt.want {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}

	t.Run("truncatesLongName", func(t *testing.T) {
		long := make([]byte, 300)
		for i := range long {
			long[i] = 'a'
		}
		got := sanitizeName(string(long))
		if len(got) != 128 {
			t.Errorf("length = %d, want 128", len(got))
		}
	})

	t.Run("truncatesOnRuneBoundary", func(t *testing.T) {
		// 50 three-byte runes = 150 bytes; a plain byte cut at 128 lands mid-rune,
		// so the whole split rune is dropped and 42 runes (126 bytes) remain
		got := sanitizeName(strings.Repeat("日", 50))
		if want := strings.Repeat("日", 42); got != want {
			t.Errorf("sanitizeName = %q (%d bytes), want %q (%d bytes)", got, len(got), want, len(want))
		}
	})
}

// testMasterKey is a fixed 32-byte key. The storage tests seal and re-open with
// the same one, which a freshly generated key per call would not allow.
func testMasterKey() []byte {
	key := make([]byte, envelope.KeySize)
	for i := range key {
		key[i] = byte(i + 1)
	}
	return key
}

// TestSealSizeless covers the one-pass padded seal for a source that declares
// no length: it is written straight into the stored file, and the frame chunk 0
// carries is filled in once the source ends.
func TestSealSizeless(t *testing.T) {
	dir := t.TempDir()
	fileID := strings.Repeat("a", 32)
	payload := []byte("an upload sealed without a declared length")

	path, err := encryptedFilePath(dir, fileID)
	if err != nil {
		t.Fatalf("encryptedFilePath error: %v", err)
	}

	up, err := sealSizeless(testMasterKey(), path, fileID, bytes.NewReader(payload), testMaxFileSize)
	if err != nil {
		t.Fatalf("sealSizeless error: %v", err)
	}
	if up.size != int64(len(payload)) {
		t.Errorf("size = %d, want %d", up.size, len(payload))
	}

	sealed, err := os.ReadFile(path) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if sealed[0] != 0x81 {
		t.Errorf("first byte = %#x, want 0x81 (stream header)", sealed[0])
	}
	if bytes.Contains(sealed, payload) {
		t.Error("plaintext found in the sealed file")
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("mode = %#o, want 0600", perm)
	}
	// chunk 0 is written last, at its own offset, so the file has to end up the
	// size the format states rather than one chunk short or one chunk long
	if want := sealedSize(int64(len(payload))); info.Size() != want {
		t.Errorf("sealed size = %d, want %d", info.Size(), want)
	}

	// it opens with the stock envelope reader, padding and all
	out := filepath.Join(t.TempDir(), "recovered")
	n, err := scheme.OpenPaddedFileAAD(testMasterKey(), out, path, []byte(fileID))
	if err != nil {
		t.Fatalf("OpenPaddedFileAAD error: %v", err)
	}
	if n != int64(len(payload)) {
		t.Errorf("recovered %d bytes, want %d", n, len(payload))
	}
	got, err := os.ReadFile(out) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("recovered %q, want %q", got, payload)
	}
}

// TestSealSizeless_MultiChunk is the case where chunk 0 is held back while
// whole chunks are sealed behind it: the payload has to come back in order, with
// nothing overwritten by the chunk written last.
func TestSealSizeless_MultiChunk(t *testing.T) {
	dir := t.TempDir()
	fileID := strings.Repeat("d", 32)

	payload := make([]byte, 2*fileChunkSize+1234)
	for i := range payload {
		payload[i] = byte(i % 251)
	}

	path, _ := encryptedFilePath(dir, fileID)
	up, err := sealSizeless(testMasterKey(), path, fileID, bytes.NewReader(payload), testMaxFileSize)
	if err != nil {
		t.Fatalf("sealSizeless error: %v", err)
	}
	if up.size != int64(len(payload)) {
		t.Fatalf("size = %d, want %d", up.size, len(payload))
	}

	out := filepath.Join(t.TempDir(), "recovered")
	if _, err := scheme.OpenPaddedFileAAD(testMasterKey(), out, path, []byte(fileID)); err != nil {
		t.Fatalf("OpenPaddedFileAAD error: %v", err)
	}
	got, err := os.ReadFile(out) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Error("recovered content differs from the payload")
	}
}

// TestSealSizeless_PaddedOpenerOnly is the format tag doing its job: a sizeless
// seal writes the padded format, so the plain reader must refuse it.
func TestSealSizeless_PaddedOpenerOnly(t *testing.T) {
	dir := t.TempDir()
	fileID := strings.Repeat("e", 32)
	path, _ := encryptedFilePath(dir, fileID)

	if _, err := sealSizeless(testMasterKey(), path, fileID, strings.NewReader("padded, not plain"), testMaxFileSize); err != nil {
		t.Fatalf("sealSizeless error: %v", err)
	}

	f, err := os.Open(path) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("open error: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	r, err := scheme.OpenReaderAAD(testMasterKey(), f, []byte(fileID))
	if err == nil {
		_, err = io.Copy(io.Discard, r)
	}
	if !errors.Is(err, envelope.ErrStreamAuth) {
		t.Errorf("err = %v, want envelope.ErrStreamAuth", err)
	}
}

func TestSealSizeless_EmptyLeavesNothing(t *testing.T) {
	dir := t.TempDir()
	fileID := strings.Repeat("f", 32)
	path, _ := encryptedFilePath(dir, fileID)

	if _, err := sealSizeless(testMasterKey(), path, fileID, bytes.NewReader(nil), testMaxFileSize); !errors.Is(err, ErrEmptyUpload) {
		t.Fatalf("err = %v, want ErrEmptyUpload", err)
	}
	// an empty upload is only recognized after the seal, so the file it sealed
	// has to be taken back off disk
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Error("a sealed empty upload was left on disk")
	}
}

func TestSealSizeless_TooLargeLeavesNothing(t *testing.T) {
	const limit = 1 << 10
	dir := t.TempDir()
	fileID := strings.Repeat("0", 32)
	path, _ := encryptedFilePath(dir, fileID)

	if _, err := sealSizeless(testMasterKey(), path, fileID, strings.NewReader(strings.Repeat("x", limit+1)), limit); !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("err = %v, want ErrFileTooLarge", err)
	}
	if entries, _ := os.ReadDir(dir); len(entries) != 0 {
		t.Errorf("directory has %d entries, want 0 (partial file left behind)", len(entries))
	}
}

// TestSealSizeless_SourceErrorLeavesNothing covers a client that goes away
// part-way. The stored file must not be left holding a fragment a download
// could ask for.
func TestSealSizeless_SourceErrorLeavesNothing(t *testing.T) {
	dir := t.TempDir()
	fileID := strings.Repeat("1", 32)
	path, _ := encryptedFilePath(dir, fileID)

	broken := io.MultiReader(
		bytes.NewReader(bytes.Repeat([]byte("x"), fileChunkSize+7)),
		errReader{err: errors.New("connection reset")},
	)
	if _, err := sealSizeless(testMasterKey(), path, fileID, broken, testMaxFileSize); err == nil {
		t.Fatal("a source that quit part-way was sealed without error")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Error("a partial stored file was left behind")
	}
}

// TestSealSizeless_RefusesExistingFile guards the O_EXCL: a file id that somehow
// repeats must never overwrite the ciphertext already stored under it.
func TestSealSizeless_RefusesExistingFile(t *testing.T) {
	dir := t.TempDir()
	fileID := strings.Repeat("c", 32)
	path, _ := encryptedFilePath(dir, fileID)

	if _, err := sealSizeless(testMasterKey(), path, fileID, strings.NewReader("the file that is already there"), testMaxFileSize); err != nil {
		t.Fatalf("sealSizeless error: %v", err)
	}
	sealed, err := os.ReadFile(path) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("read error: %v", err)
	}

	if _, err := sealSizeless(testMasterKey(), path, fileID, strings.NewReader("overwrite me"), testMaxFileSize); !errors.Is(err, os.ErrExist) {
		t.Fatalf("err = %v, want os.ErrExist", err)
	}

	after, err := os.ReadFile(path) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if !bytes.Equal(after, sealed) {
		t.Error("the stored ciphertext was modified by a refused seal")
	}
}

// TestSealSizedAndSizeless_PadAlike keeps the two padded entry points honest:
// declared length or not, the same payload has to leave the same bytes on disk,
// or the file would advertise which route it took.
func TestSealSizedAndSizeless_PadAlike(t *testing.T) {
	dir := t.TempDir()
	payload := []byte(strings.Repeat("the same payload, sealed two ways. ", 700))

	sizedID, sizelessID := strings.Repeat("2", 32), strings.Repeat("3", 32)
	sizedPath, _ := encryptedFilePath(dir, sizedID)
	sizelessPath, _ := encryptedFilePath(dir, sizelessID)

	if _, err := sealSized(testMasterKey(), sizedPath, sizedID, bytes.NewReader(payload), int64(len(payload)), testMaxFileSize); err != nil {
		t.Fatalf("sealSized error: %v", err)
	}
	if _, err := sealSizeless(testMasterKey(), sizelessPath, sizelessID, bytes.NewReader(payload), testMaxFileSize); err != nil {
		t.Fatalf("sealSizeless error: %v", err)
	}

	sizedInfo, err := os.Stat(sizedPath)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	sizelessInfo, err := os.Stat(sizelessPath)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if sizedInfo.Size() != sizelessInfo.Size() {
		t.Errorf("sized sealed to %d bytes, sizeless to %d", sizedInfo.Size(), sizelessInfo.Size())
	}
	if want := sealedSize(int64(len(payload))); sizedInfo.Size() != want {
		t.Errorf("sealed size = %d, want %d", sizedInfo.Size(), want)
	}
}
