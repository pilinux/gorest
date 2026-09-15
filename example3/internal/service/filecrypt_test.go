package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/pilinux/crypt/envelope"

	"github.com/pilinux/gorest/example3/internal/database/model"
)

const (
	// testMaxFileSize is the upload limit used by the file service under test.
	testMaxFileSize = 4 << 20

	// streamHeaderSize is the cleartext header every sealed stream starts with:
	// version(1) || saltLen(1) || salt(16) || chunkSize(4) || noncePrefix(15).
	streamHeaderSize = envelope.StreamHeaderSize

	// tagSize is the Poly1305 tag every sealed chunk carries.
	tagSize = envelope.TagSize
)

// newFileCryptService wires a FileCryptService with a loaded key, an in-memory
// metadata store and a throwaway base directory.
func newFileCryptService(t *testing.T) (*FileCryptService, *fakeFileStore, string) {
	t.Helper()
	store := newFakeFileStore()
	baseDir := t.TempDir()
	svc := NewFileCryptService(loadedKeys(t), store, baseDir, testMaxFileSize)
	return svc, store, baseDir
}

// encrypt stores content through the service and returns the created record.
func encrypt(t *testing.T, svc *FileCryptService, name string, content []byte) *model.FileRecord {
	t.Helper()
	resp, code := svc.EncryptAndStore(context.Background(), name, bytes.NewReader(content), SizeUnknown)
	if code != http.StatusCreated {
		t.Fatalf("EncryptAndStore status = %d, want 201 (resp: %v)", code, resp.Message)
	}
	rec, ok := resp.Message.(*model.FileRecord)
	if !ok {
		t.Fatalf("unexpected message type %T", resp.Message)
	}
	return rec
}

// download opens fileID the way OpenForDownload does and drains the payload
// into a buffer. It goes through openSealed, which returns what went wrong
// instead of a status, so a failure at open (before a status would be sent)
// and one while streaming both come back as err.
func download(t *testing.T, svc *FileCryptService, fileID string) (*Download, []byte, error) {
	t.Helper()
	rec, err := svc.store.GetByFileID(context.Background(), fileID)
	if err != nil {
		t.Fatalf("GetByFileID error: %v", err)
	}
	dl, err := svc.openSealed(rec)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		_ = dl.Close()
	}()

	var buf bytes.Buffer
	_, err = dl.WriteTo(&buf)
	return dl, buf.Bytes(), err
}

// randomContent returns n bytes of randomness, so scanning a ciphertext for
// leftover plaintext is meaningful.
func randomContent(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatalf("rand.Read error: %v", err)
	}
	return b
}

// sealedSize is the on-disk size of a sealed stream carrying an n-byte payload:
// the header, the padded plaintext and one tag per chunk.
func sealedSize(n int64) int64 {
	padded := envelope.PaddedSize(n)
	chunks := (padded + fileChunkSize - 1) / fileChunkSize
	return streamHeaderSize + padded + tagSize*chunks
}

// storedFiles counts the sealed files in baseDir and fails if anything else was
// left behind: an upload is sealed straight into its stored file, so the
// directory holds nothing but finished ciphertext.
func storedFiles(t *testing.T, baseDir string) int {
	t.Helper()

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}
	sealed := 0
	for _, e := range entries {
		if e.IsDir() {
			t.Errorf("unexpected directory %q under the encrypted-files directory", e.Name())
			continue
		}
		if filepath.Ext(e.Name()) != encryptedFileExt {
			t.Errorf("unexpected file %q under the encrypted-files directory", e.Name())
			continue
		}
		sealed++
	}
	return sealed
}

func TestFileCrypt_EncryptStoreDecrypt(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	content := []byte("top secret file content\n")

	rec := encrypt(t, svc, "notes/secret.txt", content)
	if rec.Name != "secret.txt" {
		t.Errorf("name = %q, want secret.txt", rec.Name)
	}
	if rec.Size != int64(len(content)) {
		t.Errorf("size = %d, want %d", rec.Size, len(content))
	}
	if _, saved := store.records[rec.FileID]; !saved {
		t.Error("metadata was not stored")
	}

	// the ciphertext file must exist on disk, carry the streaming format and
	// hold none of the plaintext
	path, _ := encryptedFilePath(baseDir, rec.FileID)
	onDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("encrypted file missing on disk: %v", err)
	}
	if onDisk[0] != 0x81 {
		t.Errorf("version byte = %#x, want 0x81 (stream format)", onDisk[0])
	}
	if bytes.Contains(onDisk, content) {
		t.Error("plaintext found inside the encrypted file")
	}
	if got, want := int64(len(onDisk)), sealedSize(int64(len(content))); got != want {
		t.Errorf("sealed size = %d, want %d", got, want)
	}
	// the payload was padded, so the file is longer than the plaintext plus a
	// fixed overhead
	if int64(len(onDisk)) <= int64(len(content))+streamHeaderSize+tagSize {
		t.Error("file is not padded")
	}

	// exactly one sealed file, and no plaintext left in the incoming directory
	if n := storedFiles(t, baseDir); n != 1 {
		t.Errorf("stored %d files, want 1", n)
	}

	// decrypt returns the original content and display name
	dl, data, err := download(t, svc, rec.FileID)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Error("decrypted content does not match original")
	}
	if dl.Name != "secret.txt" {
		t.Errorf("name = %q, want secret.txt", dl.Name)
	}
	if dl.Size != int64(len(content)) {
		t.Errorf("size = %d, want %d", dl.Size, len(content))
	}
}

func TestFileCrypt_MultiChunkRoundTrip(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	// two full chunks plus a partial one, so the chunk counter and the
	// final-chunk flag are both exercised
	content := randomContent(t, 2*fileChunkSize+512)

	rec := encrypt(t, svc, "big.bin", content)
	if rec.Size != int64(len(content)) {
		t.Errorf("size = %d, want %d", rec.Size, len(content))
	}

	path, _ := encryptedFilePath(baseDir, rec.FileID)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if got, want := info.Size(), sealedSize(int64(len(content))); got != want {
		t.Errorf("sealed size = %d, want %d (3 chunks)", got, want)
	}

	_, data, err := download(t, svc, rec.FileID)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Error("decrypted content does not match original")
	}
}

// scanForPlaintext fails if any file under baseDir holds marker, which is a
// slice of the payload being uploaded.
func scanForPlaintext(t *testing.T, baseDir string, marker []byte) {
	t.Helper()

	err := filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(path) // #nosec G304 -- test-owned temp path
		if err != nil {
			return err
		}
		if bytes.Contains(data, marker) {
			t.Errorf("plaintext found on disk in %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir error: %v", err)
	}
}

// TestFileCrypt_NoPlaintextOnDiskDuringUpload looks at the disk while an
// upload of undeclared length is open and only part written. The only thing
// being written is the stored file, and every byte of it is already sealed: the
// payload never touches the disk in the clear, and nothing is staged beside it.
func TestFileCrypt_NoPlaintextOnDiskDuringUpload(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)

	// several chunks, so the stored file is genuinely part-written when we look
	content := randomContent(t, 3*fileChunkSize)
	marker := content[:64] // a recognizable slice of the payload

	var looked bool
	src := &hookReader{
		r:  bytes.NewReader(content),
		at: 2 * fileChunkSize,
		fn: func() {
			looked = true
			scanForPlaintext(t, baseDir, marker)
		},
	}

	if _, code := svc.EncryptAndStore(context.Background(), "secret.bin", src, SizeUnknown); code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", code)
	}
	if !looked {
		t.Fatal("the mid-upload check never ran")
	}

	// and nothing is left in the clear once the file is stored either
	scanForPlaintext(t, baseDir, marker)
}

func TestFileCrypt_EncryptEmpty(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)

	resp, code := svc.EncryptAndStore(context.Background(), "empty.txt", strings.NewReader(""), SizeUnknown)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (resp: %v)", code, resp.Message)
	}
	if n := storedFiles(t, baseDir); n != 0 {
		t.Errorf("stored %d files, want 0", n)
	}
}

func TestFileCrypt_EncryptTooLarge(t *testing.T) {
	const limit = 1 << 12
	baseDir := t.TempDir()
	svc := NewFileCryptService(loadedKeys(t), newFakeFileStore(), baseDir, limit)
	ctx := context.Background()

	// exactly at the limit is still accepted
	if _, code := svc.EncryptAndStore(ctx, "fits.bin", bytes.NewReader(randomContent(t, limit)), SizeUnknown); code != http.StatusCreated {
		t.Errorf("status at the limit = %d, want 201", code)
	}

	// one byte more is refused, and nothing is left behind
	if _, code := svc.EncryptAndStore(ctx, "toobig.bin", bytes.NewReader(randomContent(t, limit+1)), SizeUnknown); code != http.StatusRequestEntityTooLarge {
		t.Errorf("status past the limit = %d, want 413", code)
	}
	if n := storedFiles(t, baseDir); n != 1 {
		t.Errorf("stored %d files, want 1 (partial file left behind)", n)
	}
}

func TestFileCrypt_EncryptReadErrorLeavesNothing(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	// a client that disappears mid-upload
	broken := io.MultiReader(bytes.NewReader([]byte("partial")), errReader{err: errors.New("connection reset")})

	if _, code := svc.EncryptAndStore(context.Background(), "broken.bin", broken, SizeUnknown); code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", code)
	}
	if n := storedFiles(t, baseDir); n != 0 {
		t.Errorf("stored %d files, want 0", n)
	}
	if len(store.records) != 0 {
		t.Errorf("stored %d records, want 0", len(store.records))
	}
}

func TestFileCrypt_MaxSize(t *testing.T) {
	const limit = 4 << 20
	svc := NewFileCryptService(loadedKeys(t), newFakeFileStore(), t.TempDir(), limit)
	if got := svc.MaxSize(); got != limit {
		t.Errorf("MaxSize() = %d, want %d", got, limit)
	}
}

// TestFileCrypt_BodyLimitIsClientError covers the cap the handler puts on the
// whole request body. It runs out while the file is being read, so the failure
// arrives here as an http.MaxBytesError rather than as ErrFileTooLarge, and it
// is still the client's error, not a 500.
func TestFileCrypt_BodyLimitIsClientError(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	capped := io.MultiReader(
		bytes.NewReader([]byte("the part that fit")),
		errReader{err: &http.MaxBytesError{Limit: 17}},
	)

	resp, code := svc.EncryptAndStore(context.Background(), "capped.bin", capped, SizeUnknown)
	if code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413", code)
	}
	if msg, ok := resp.Message.(string); !ok || msg != "request body is too large" {
		t.Errorf("message = %v, want \"request body is too large\"", resp.Message)
	}
	if n := storedFiles(t, baseDir); n != 0 {
		t.Errorf("stored %d files, want 0", n)
	}
	if len(store.records) != 0 {
		t.Errorf("stored %d records, want 0", len(store.records))
	}
}

func TestFileCrypt_DecryptInvalidID(t *testing.T) {
	svc, _, _ := newFileCryptService(t)
	if _, _, code := svc.OpenForDownload(context.Background(), "../evil"); code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", code)
	}
}

func TestFileCrypt_DecryptNotFound(t *testing.T) {
	svc, _, _ := newFileCryptService(t)
	// well-formed id that was never stored
	missing := strings.Repeat("a", 32)
	if _, _, code := svc.OpenForDownload(context.Background(), missing); code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", code)
	}
}

func TestFileCrypt_TamperedCiphertextRejected(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	rec := encrypt(t, svc, "tampered.txt", []byte("authentic content"))
	path, _ := encryptedFilePath(baseDir, rec.FileID)

	blob, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	// flip a bit inside the first chunk, past the header
	blob[streamHeaderSize+1] ^= 0x01
	if err := os.WriteFile(path, blob, 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	if _, _, err := download(t, svc, rec.FileID); err == nil {
		t.Error("tampered ciphertext was accepted")
	}
}

func TestFileCrypt_TruncatedCiphertextRejected(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	rec := encrypt(t, svc, "cut.txt", []byte("a whole authentic file"))
	path, _ := encryptedFilePath(baseDir, rec.FileID)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	if err := os.Truncate(path, info.Size()-1); err != nil {
		t.Fatalf("truncate error: %v", err)
	}

	if _, _, err := download(t, svc, rec.FileID); err == nil {
		t.Error("truncated ciphertext was accepted")
	}
}

func TestFileCrypt_CiphertextBoundToFileID(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	rec := encrypt(t, svc, "bound.txt", []byte("belongs to one id only"))

	// move the sealed bytes to a second id, metadata and all
	otherID := strings.Repeat("b", 32)
	srcPath, _ := encryptedFilePath(baseDir, rec.FileID)
	dstPath, _ := encryptedFilePath(baseDir, otherID)
	blob, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if err := os.WriteFile(dstPath, blob, 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}
	moved := *rec
	moved.FileID = otherID
	store.records[otherID] = &moved

	// the file id is the AAD of every chunk, so the copy no longer opens
	if _, _, err := download(t, svc, otherID); err == nil {
		t.Error("ciphertext opened under a foreign file id")
	}
}

func TestFileCrypt_DigestMismatchDetected(t *testing.T) {
	svc, store, _ := newFileCryptService(t)
	rec := encrypt(t, svc, "swapped.txt", []byte("content the record no longer describes"))

	// the ciphertext still authenticates; only the recorded digest disagrees
	store.records[rec.FileID].Sha256 = strings.Repeat("0", 64)

	if _, _, err := download(t, svc, rec.FileID); !errors.Is(err, ErrIntegrityCheckFailed) {
		t.Errorf("err = %v, want ErrIntegrityCheckFailed", err)
	}
}

func TestFileCrypt_Delete(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	ctx := context.Background()

	rec := encrypt(t, svc, "gone.txt", []byte("bye"))
	path, _ := encryptedFilePath(baseDir, rec.FileID)

	if _, code := svc.DeleteStored(ctx, rec.FileID); code != http.StatusOK {
		t.Fatalf("DeleteStored status = %d, want 200", code)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Error("encrypted file still on disk after delete")
	}

	// deleting again must report not found
	if _, code := svc.DeleteStored(ctx, rec.FileID); code != http.StatusNotFound {
		t.Errorf("second delete status = %d, want 404", code)
	}
}

func TestFileCrypt_CreateErrorRollsBackFile(t *testing.T) {
	store := newFakeFileStore()
	store.createErr = errors.New("db down")
	baseDir := t.TempDir()
	svc := NewFileCryptService(loadedKeys(t), store, baseDir, testMaxFileSize)

	if _, code := svc.EncryptAndStore(context.Background(), "x.txt", strings.NewReader("data"), SizeUnknown); code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", code)
	}

	// the on-disk ciphertext must have been rolled back, leaving no files
	if n := storedFiles(t, baseDir); n != 0 {
		t.Errorf("stored %d files, want 0 (rollback failed)", n)
	}
}

func TestFileCrypt_PaddingHidesLength(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)

	// three payloads that share one Padmé bucket have to end up as files of
	// exactly the same size, so the length no longer identifies the document
	sizes := []int{94500, 96037, 96200}
	var bucketSize int64

	for _, size := range sizes {
		rec := encrypt(t, svc, "doc.bin", randomContent(t, size))
		// the true length still travels in the record, not in the file
		if rec.Size != int64(size) {
			t.Errorf("recorded size = %d, want %d", rec.Size, size)
		}

		path, _ := encryptedFilePath(baseDir, rec.FileID)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat error: %v", err)
		}
		if bucketSize == 0 {
			bucketSize = info.Size()
			continue
		}
		if info.Size() != bucketSize {
			t.Errorf("a %d-byte payload sealed to %d bytes, want %d (length leaks)", size, info.Size(), bucketSize)
		}
	}

	// and the bucket is the one the envelope package promises
	if want := sealedSize(int64(sizes[0])); bucketSize != want {
		t.Errorf("bucket size = %d, want %d", bucketSize, want)
	}
	// the overhead Padmé costs stays modest
	if overhead := float64(bucketSize-int64(sizes[0])) / float64(sizes[0]); overhead > 0.12 {
		t.Errorf("padding overhead = %.1f%%, want at most 12%%", overhead*100)
	}
}

func TestFileCrypt_PaddedRoundTripAcrossSizes(t *testing.T) {
	svc, _, _ := newFileCryptService(t)

	// sizes around the chunk boundary: the padding may fill out the final chunk
	// or add whole chunks after it
	for _, size := range []int{1, 8, fileChunkSize - 1, fileChunkSize, fileChunkSize + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			content := randomContent(t, size)
			rec := encrypt(t, svc, "sized.bin", content)

			_, data, err := download(t, svc, rec.FileID)
			if err != nil {
				t.Fatalf("WriteTo error: %v", err)
			}
			if !bytes.Equal(data, content) {
				t.Errorf("round trip of %d bytes returned %d bytes of different content", size, len(data))
			}
		})
	}
}

func TestFileCrypt_UnpaddedStreamRejected(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	content := []byte("sealed by the plain stream API, without a padded frame")

	// seal a well-formed but unpadded stream under a valid id and record. Both
	// sealers take the file id as the AAD; the envelope package binds a
	// different format tag in beside it for each.
	fileID := strings.Repeat("c", 32)
	path, _ := encryptedFilePath(baseDir, fileID)
	masterKey, err := svc.keys.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey error: %v", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	if _, err := scheme.SealStreamAAD(masterKey, f, bytes.NewReader(content), []byte(fileID)); err != nil {
		t.Fatalf("SealStreamAAD error: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	store.records[fileID] = &model.FileRecord{
		FileID: fileID,
		Name:   "unpadded.bin",
		Size:   int64(len(content)),
		Sha256: envelope.Sha256Hex(content),
	}

	// padded and plain streams are domain-separated by their AAD, so the
	// reader never even gets as far as looking for a frame
	if _, _, err := download(t, svc, fileID); !errors.Is(err, envelope.ErrStreamAuth) {
		t.Errorf("err = %v, want envelope.ErrStreamAuth", err)
	}
}

func TestFileCrypt_RecordSizeTamperingRejected(t *testing.T) {
	for _, delta := range []int64{-1, +1} {
		t.Run(strconv.FormatInt(delta, 10), func(t *testing.T) {
			svc, store, _ := newFileCryptService(t)
			rec := encrypt(t, svc, "resized.bin", []byte("a payload of one particular length"))

			// the length inside the sealed plaintext is authenticated, so a
			// record that claims another one is refused at open, before any
			// byte goes out
			store.records[rec.FileID].Size = rec.Size + delta

			if _, _, err := download(t, svc, rec.FileID); !errors.Is(err, ErrIntegrityCheckFailed) {
				t.Errorf("err = %v, want ErrIntegrityCheckFailed", err)
			}
		})
	}
}

// TestFileCrypt_PaddedFailuresRefusedBeforeStatus is what opening through the
// frame buys: for a padded file, a damaged first chunk, a foreign id, the wrong
// format or a record whose size disagrees are all caught while a status can
// still say so, not after a 200 has gone out.
func TestFileCrypt_PaddedFailuresRefusedBeforeStatus(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, svc *FileCryptService, store *fakeFileStore, baseDir string) string
	}{
		{
			name: "flippedBit",
			setup: func(t *testing.T, svc *FileCryptService, _ *fakeFileStore, baseDir string) string {
				rec := encrypt(t, svc, "tampered.txt", []byte("authentic content"))
				path, _ := encryptedFilePath(baseDir, rec.FileID)
				blob, err := os.ReadFile(path) // #nosec G304 -- test-owned temp path
				if err != nil {
					t.Fatalf("read error: %v", err)
				}
				blob[streamHeaderSize+1] ^= 0x01
				if err := os.WriteFile(path, blob, 0o600); err != nil {
					t.Fatalf("write error: %v", err)
				}
				return rec.FileID
			},
		},
		{
			name: "recordSize",
			setup: func(t *testing.T, svc *FileCryptService, store *fakeFileStore, _ string) string {
				rec := encrypt(t, svc, "resized.bin", []byte("a payload of one particular length"))
				store.records[rec.FileID].Size++
				return rec.FileID
			},
		},
		{
			name: "plainFileFlaggedPadded",
			setup: func(t *testing.T, svc *FileCryptService, store *fakeFileStore, _ string) string {
				rec := encryptUnpadded(t, svc, "plain.bin", []byte("sealed with no padding at all"))
				store.records[rec.FileID].Unpadded = false
				return rec.FileID
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, store, baseDir := newFileCryptService(t)
			fileID := tc.setup(t, svc, store, baseDir)

			dl, resp, code := svc.OpenForDownload(context.Background(), fileID)
			if code != http.StatusInternalServerError {
				t.Errorf("status = %d, want 500 (resp: %v)", code, resp.Message)
			}
			if dl != nil {
				_ = dl.Close()
				t.Error("a download was handed out for a file that failed to open")
			}
		})
	}
}

// TestFileCrypt_DecryptFileGoneFromDisk: a record whose ciphertext has vanished
// is a missing file, not a server error.
func TestFileCrypt_DecryptFileGoneFromDisk(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	rec := encrypt(t, svc, "gone.txt", []byte("soon to be missing"))
	path, _ := encryptedFilePath(baseDir, rec.FileID)
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove error: %v", err)
	}

	if _, _, code := svc.OpenForDownload(context.Background(), rec.FileID); code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", code)
	}
}

// TestFileCrypt_WideChunksRefused: the header, and the chunk width it names, is
// read before anything authenticates, so example3 accepts no chunk wider than
// it writes. A file naming wider ones is refused before a buffer that wide is
// allocated, even though the key, the labels and the AAD all fit.
func TestFileCrypt_WideChunksRefused(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	content := []byte("sealed with chunks wider than example3 accepts")

	wide := envelope.New(envelope.Config{
		KEKLabel:    kekLabel,
		SubKeyLabel: subKeyLabel,
		ChunkSize:   2 * maxAcceptedChunkSize,
	})
	fileID := strings.Repeat("e", 32)
	path, _ := encryptedFilePath(baseDir, fileID)
	masterKey, err := svc.keys.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey error: %v", err)
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	if _, err := wide.SealPaddedStreamAAD(masterKey, f, bytes.NewReader(content), int64(len(content)), []byte(fileID)); err != nil {
		t.Fatalf("SealPaddedStreamAAD error: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close error: %v", err)
	}
	store.records[fileID] = &model.FileRecord{
		FileID: fileID,
		Name:   "wide.bin",
		Size:   int64(len(content)),
		Sha256: envelope.Sha256Hex(content),
	}

	if _, _, err := download(t, svc, fileID); !errors.Is(err, envelope.ErrInvalidChunkSize) {
		t.Errorf("err = %v, want envelope.ErrInvalidChunkSize", err)
	}
}

// encryptSized stores content through the single-pass path, declaring its
// length the way a raw upload's Content-Length does.
func encryptSized(t *testing.T, svc *FileCryptService, name string, content []byte) *model.FileRecord {
	t.Helper()
	resp, code := svc.EncryptAndStore(context.Background(), name, bytes.NewReader(content), int64(len(content)))
	if code != http.StatusCreated {
		t.Fatalf("EncryptAndStore status = %d, want 201 (resp: %v)", code, resp.Message)
	}
	rec, ok := resp.Message.(*model.FileRecord)
	if !ok {
		t.Fatalf("unexpected message type %T", resp.Message)
	}
	return rec
}

// TestFileCrypt_SizedUploadWritesOnlyTheStoredFile is what a declared length
// buys: the padding is sized before the first chunk, so the upload is sealed
// into its final file in order, with nothing held back and nothing else on disk.
func TestFileCrypt_SizedUploadWritesOnlyTheStoredFile(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	content := randomContent(t, 2*fileChunkSize+77)
	marker := content[:64] // a recognizable slice of the payload

	// look at the disk while the upload is only part written: the sealed file
	// is the only thing being written, and it is ciphertext
	var looked bool
	src := &hookReader{
		r:  bytes.NewReader(content),
		at: fileChunkSize,
		fn: func() {
			looked = true
			scanForPlaintext(t, baseDir, marker)
		},
	}

	resp, code := svc.EncryptAndStore(context.Background(), "declared.bin", src, int64(len(content)))
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (resp: %v)", code, resp.Message)
	}
	if !looked {
		t.Fatal("the mid-upload check never ran")
	}
	rec, ok := resp.Message.(*model.FileRecord)
	if !ok {
		t.Fatalf("unexpected message type %T", resp.Message)
	}
	if rec.Size != int64(len(content)) {
		t.Errorf("size = %d, want %d", rec.Size, len(content))
	}
	if n := storedFiles(t, baseDir); n != 1 {
		t.Errorf("stored %d files, want 1", n)
	}

	_, data, err := download(t, svc, rec.FileID)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Error("decrypted content does not match original")
	}
	scanForPlaintext(t, baseDir, marker)
}

// TestFileCrypt_SizedAndSizelessPadAlike keeps the two padded paths honest:
// whether the length was declared or counted while the upload streamed by, the
// same payload has to end up as the same number of bytes on disk. Otherwise the
// file would advertise which route it took.
func TestFileCrypt_SizedAndSizelessPadAlike(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	content := randomContent(t, 94500)

	sizedOnDisk := storedSize(t, baseDir, encryptSized(t, svc, "declared.bin", content).FileID)
	sizelessOnDisk := storedSize(t, baseDir, encrypt(t, svc, "undeclared.bin", content).FileID)

	if sizedOnDisk != sizelessOnDisk {
		t.Errorf("a declared length sealed to %d bytes, an undeclared one to %d", sizedOnDisk, sizelessOnDisk)
	}
	if want := sealedSize(int64(len(content))); sizedOnDisk != want {
		t.Errorf("sealed size = %d, want %d", sizedOnDisk, want)
	}
}

// storedSize is the on-disk size of the sealed file for fileID.
func storedSize(t *testing.T, baseDir, fileID string) int64 {
	t.Helper()
	path, err := encryptedFilePath(baseDir, fileID)
	if err != nil {
		t.Fatalf("encryptedFilePath error: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	return info.Size()
}

// TestFileCrypt_SizedUploadWrongLength covers a client whose body does not
// match the length it declared. The mismatch is caught at the payload boundary,
// before any padding is written, and it is the client's error.
func TestFileCrypt_SizedUploadWrongLength(t *testing.T) {
	content := []byte("thirty-two bytes of file content")

	tests := []struct {
		name     string
		declared int64
		message  string
	}{
		{name: "shortSource", declared: int64(len(content)) + 8, message: "upload ended before its declared length"},
		{name: "longSource", declared: int64(len(content)) - 8, message: "upload is longer than its declared length"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc, store, baseDir := newFileCryptService(t)

			resp, code := svc.EncryptAndStore(context.Background(), "lying.bin", bytes.NewReader(content), tc.declared)
			if code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400 (resp: %v)", code, resp.Message)
			}
			// which way it missed is the client's to know
			if msg, ok := resp.Message.(string); !ok || msg != tc.message {
				t.Errorf("message = %v, want %q", resp.Message, tc.message)
			}
			if n := storedFiles(t, baseDir); n != 0 {
				t.Errorf("stored %d files, want 0", n)
			}
			if len(store.records) != 0 {
				t.Errorf("stored %d records, want 0", len(store.records))
			}
		})
	}
}

// TestFileCrypt_SizedUploadTooLarge is the other thing a declared length buys:
// an oversized upload is refused without reading a byte of it.
func TestFileCrypt_SizedUploadTooLarge(t *testing.T) {
	const limit = 1 << 12
	baseDir := t.TempDir()
	svc := NewFileCryptService(loadedKeys(t), newFakeFileStore(), baseDir, limit)

	unread := errReader{err: errors.New("the body must not be read")}
	resp, code := svc.EncryptAndStore(context.Background(), "toobig.bin", unread, limit+1)
	if code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413 (resp: %v)", code, resp.Message)
	}
	if n := storedFiles(t, baseDir); n != 0 {
		t.Errorf("stored %d files, want 0", n)
	}
}

func TestFileCrypt_SizedUploadEmpty(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)

	resp, code := svc.EncryptAndStore(context.Background(), "empty.txt", bytes.NewReader(nil), 0)
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (resp: %v)", code, resp.Message)
	}
	if n := storedFiles(t, baseDir); n != 0 {
		t.Errorf("stored %d files, want 0", n)
	}
}

// TestFileCrypt_SizedUploadOpensWithStockReader is the interoperability claim:
// a single-pass file is an ordinary padded stream, so the envelope package's
// own opener reads it back.
func TestFileCrypt_SizedUploadOpensWithStockReader(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	content := randomContent(t, 5000)
	rec := encryptSized(t, svc, "stock.bin", content)

	path, _ := encryptedFilePath(baseDir, rec.FileID)
	masterKey, err := svc.keys.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey error: %v", err)
	}
	out := filepath.Join(t.TempDir(), "recovered")
	if _, err := scheme.OpenPaddedFileAAD(masterKey, out, path, []byte(rec.FileID)); err != nil {
		t.Fatalf("OpenPaddedFileAAD error: %v", err)
	}
	got, err := os.ReadFile(out) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Error("stock reader recovered different content")
	}
}

// plainSealedSize is the on-disk size of an unpadded sealed stream carrying an
// n-byte payload: the header, the payload itself and one tag per chunk.
func plainSealedSize(n int64) int64 {
	chunks := max((n+fileChunkSize-1)/fileChunkSize, 1)
	return streamHeaderSize + n + tagSize*chunks
}

// encryptUnpadded stores content through the plain (unpadded) path.
func encryptUnpadded(t *testing.T, svc *FileCryptService, name string, content []byte) *model.FileRecord {
	t.Helper()
	resp, code := svc.EncryptAndStoreUnpadded(context.Background(), name, bytes.NewReader(content))
	if code != http.StatusCreated {
		t.Fatalf("EncryptAndStoreUnpadded status = %d, want 201 (resp: %v)", code, resp.Message)
	}
	rec, ok := resp.Message.(*model.FileRecord)
	if !ok {
		t.Fatalf("unexpected message type %T", resp.Message)
	}
	return rec
}

// TestFileCrypt_UnpaddedRoundTrip is the plain path end to end: one pass, no
// spool, no padding, and the record says so.
func TestFileCrypt_UnpaddedRoundTrip(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	content := randomContent(t, 2*fileChunkSize+321)
	marker := content[:64]

	// look at the disk mid-upload: the sealed file is the only thing written
	var looked bool
	src := &hookReader{
		r:  bytes.NewReader(content),
		at: fileChunkSize,
		fn: func() {
			looked = true
			scanForPlaintext(t, baseDir, marker)
		},
	}

	resp, code := svc.EncryptAndStoreUnpadded(context.Background(), "plain.bin", src)
	if code != http.StatusCreated {
		t.Fatalf("status = %d, want 201 (resp: %v)", code, resp.Message)
	}
	if !looked {
		t.Fatal("the mid-upload check never ran")
	}
	rec, ok := resp.Message.(*model.FileRecord)
	if !ok {
		t.Fatalf("unexpected message type %T", resp.Message)
	}
	if !rec.Unpadded {
		t.Error("record does not say the file is unpadded")
	}
	if rec.Size != int64(len(content)) {
		t.Errorf("size = %d, want %d", rec.Size, len(content))
	}
	if want := envelope.Sha256Hex(content); rec.Sha256 != want {
		t.Errorf("sha256 = %q, want %q", rec.Sha256, want)
	}
	if _, saved := store.records[rec.FileID]; !saved {
		t.Error("metadata was not stored")
	}

	// no length was declared and none was needed, so the stored file is the
	// only thing that was written
	if n := storedFiles(t, baseDir); n != 1 {
		t.Errorf("stored %d files, want 1", n)
	}

	// the file is the payload plus the fixed stream overhead: no padding
	if got, want := storedSize(t, baseDir, rec.FileID), plainSealedSize(int64(len(content))); got != want {
		t.Errorf("sealed size = %d, want %d", got, want)
	}

	_, data, err := download(t, svc, rec.FileID)
	if err != nil {
		t.Fatalf("WriteTo error: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Error("decrypted content does not match original")
	}
	scanForPlaintext(t, baseDir, marker)
}

// TestFileCrypt_UnpaddedLeaksLength is the trade-off, stated as a test: three
// payloads a padded seal would collapse onto one size stay three sizes here.
func TestFileCrypt_UnpaddedLeaksLength(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	sizes := []int{94500, 96037, 96200}

	seen := make(map[int64]bool, len(sizes))
	for _, size := range sizes {
		rec := encryptUnpadded(t, svc, "doc.bin", randomContent(t, size))
		onDisk := storedSize(t, baseDir, rec.FileID)
		if want := plainSealedSize(int64(size)); onDisk != want {
			t.Errorf("a %d-byte payload sealed to %d bytes, want %d", size, onDisk, want)
		}
		seen[onDisk] = true
	}
	if len(seen) != len(sizes) {
		t.Errorf("%d distinct sizes on disk, want %d", len(seen), len(sizes))
	}

	// and the same payloads padded do collapse onto one
	padded := storedSize(t, baseDir, encrypt(t, svc, "doc.bin", randomContent(t, sizes[0])).FileID)
	for _, size := range sizes[1:] {
		if got := storedSize(t, baseDir, encrypt(t, svc, "doc.bin", randomContent(t, size)).FileID); got != padded {
			t.Errorf("padded %d-byte payload sealed to %d bytes, want %d", size, got, padded)
		}
	}
}

// TestFileCrypt_UnpaddedOpensWithStockReader: a plain file is an ordinary
// sealed stream, so the envelope package's own reader opens it.
func TestFileCrypt_UnpaddedOpensWithStockReader(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	content := randomContent(t, 3000)
	rec := encryptUnpadded(t, svc, "stock.bin", content)

	path, _ := encryptedFilePath(baseDir, rec.FileID)
	f, err := os.Open(path) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("open error: %v", err)
	}
	defer func() {
		_ = f.Close()
	}()

	masterKey, err := svc.keys.MasterKey()
	if err != nil {
		t.Fatalf("MasterKey error: %v", err)
	}
	var out bytes.Buffer
	if _, err := scheme.OpenStreamAAD(masterKey, &out, f, []byte(rec.FileID)); err != nil {
		t.Fatalf("OpenStreamAAD error: %v", err)
	}
	if !bytes.Equal(out.Bytes(), content) {
		t.Error("stock reader recovered different content")
	}
}

// TestFileCrypt_FormatFlagIsAuthenticated: flipping the record's format flag
// cannot make a file open the other way. The padded format's AAD tag is what
// stops it, so a flipped flag fails to authenticate rather than mis-parsing.
func TestFileCrypt_FormatFlagIsAuthenticated(t *testing.T) {
	svc, store, _ := newFileCryptService(t)

	padded := encrypt(t, svc, "padded.bin", []byte("sealed with a padding frame"))
	store.records[padded.FileID].Unpadded = true
	if _, _, err := download(t, svc, padded.FileID); !errors.Is(err, envelope.ErrStreamAuth) {
		t.Errorf("padded file read as plain: err = %v, want envelope.ErrStreamAuth", err)
	}

	plain := encryptUnpadded(t, svc, "plain.bin", []byte("sealed with no padding at all"))
	store.records[plain.FileID].Unpadded = false
	if _, _, err := download(t, svc, plain.FileID); !errors.Is(err, envelope.ErrStreamAuth) {
		t.Errorf("plain file read as padded: err = %v, want envelope.ErrStreamAuth", err)
	}
}

// TestFileCrypt_UnpaddedRecordSizeTamperingRejected: a plain stream carries no
// sealed length to check the record against, so the guard is what follows the
// payload: for an unpadded file, nothing may.
//
// A record claiming too little is caught by a one-byte probe past the payload;
// one claiming too much, by the stream ending first.
func TestFileCrypt_UnpaddedRecordSizeTamperingRejected(t *testing.T) {
	for _, delta := range []int64{-1, +1} {
		t.Run(strconv.FormatInt(delta, 10), func(t *testing.T) {
			svc, store, _ := newFileCryptService(t)
			rec := encryptUnpadded(t, svc, "resized.bin", []byte("a payload of one particular length"))

			store.records[rec.FileID].Size = rec.Size + delta

			if _, _, err := download(t, svc, rec.FileID); !errors.Is(err, ErrIntegrityCheckFailed) {
				t.Errorf("err = %v, want ErrIntegrityCheckFailed", err)
			}
		})
	}
}

func TestFileCrypt_UnpaddedEmpty(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)

	resp, code := svc.EncryptAndStoreUnpadded(context.Background(), "empty.txt", strings.NewReader(""))
	if code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400 (resp: %v)", code, resp.Message)
	}
	if n := storedFiles(t, baseDir); n != 0 {
		t.Errorf("stored %d files, want 0", n)
	}
	if len(store.records) != 0 {
		t.Errorf("stored %d records, want 0", len(store.records))
	}
}

func TestFileCrypt_UnpaddedTooLarge(t *testing.T) {
	const limit = 1 << 12
	baseDir := t.TempDir()
	svc := NewFileCryptService(loadedKeys(t), newFakeFileStore(), baseDir, limit)
	ctx := context.Background()

	// exactly at the limit is still accepted
	if _, code := svc.EncryptAndStoreUnpadded(ctx, "fits.bin", bytes.NewReader(randomContent(t, limit))); code != http.StatusCreated {
		t.Errorf("status at the limit = %d, want 201", code)
	}
	// one byte more is refused, and the partial file is removed
	if _, code := svc.EncryptAndStoreUnpadded(ctx, "toobig.bin", bytes.NewReader(randomContent(t, limit+1))); code != http.StatusRequestEntityTooLarge {
		t.Errorf("status past the limit = %d, want 413", code)
	}
	if n := storedFiles(t, baseDir); n != 1 {
		t.Errorf("stored %d files, want 1 (partial file left behind)", n)
	}
}

func TestFileCrypt_UnpaddedReadErrorLeavesNothing(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	broken := io.MultiReader(bytes.NewReader([]byte("partial")), errReader{err: errors.New("connection reset")})

	if _, code := svc.EncryptAndStoreUnpadded(context.Background(), "broken.bin", broken); code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", code)
	}
	if n := storedFiles(t, baseDir); n != 0 {
		t.Errorf("stored %d files, want 0", n)
	}
	if len(store.records) != 0 {
		t.Errorf("stored %d records, want 0", len(store.records))
	}
}

// TestFileCrypt_UnpaddedTamperedCiphertextRejected: no padding, but every chunk
// is still authenticated.
func TestFileCrypt_UnpaddedTamperedCiphertextRejected(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	rec := encryptUnpadded(t, svc, "tampered.bin", []byte("authentic content"))

	path, _ := encryptedFilePath(baseDir, rec.FileID)
	blob, err := os.ReadFile(path) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	blob[streamHeaderSize+1] ^= 0x01
	if err := os.WriteFile(path, blob, 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}

	if _, _, err := download(t, svc, rec.FileID); err == nil {
		t.Error("tampered ciphertext was accepted")
	}
}

// TestFileCrypt_UnpaddedCiphertextBoundToFileID: the file id is still the AAD.
func TestFileCrypt_UnpaddedCiphertextBoundToFileID(t *testing.T) {
	svc, store, baseDir := newFileCryptService(t)
	rec := encryptUnpadded(t, svc, "bound.bin", []byte("belongs to one id only"))

	otherID := strings.Repeat("d", 32)
	srcPath, _ := encryptedFilePath(baseDir, rec.FileID)
	dstPath, _ := encryptedFilePath(baseDir, otherID)
	blob, err := os.ReadFile(srcPath) // #nosec G304 -- test-owned temp path
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	if err := os.WriteFile(dstPath, blob, 0o600); err != nil {
		t.Fatalf("write error: %v", err)
	}
	moved := *rec
	moved.FileID = otherID
	store.records[otherID] = &moved

	if _, _, err := download(t, svc, otherID); err == nil {
		t.Error("ciphertext opened under a foreign file id")
	}
}

// TestFileCrypt_UnpaddedDelete: one delete route serves either format.
func TestFileCrypt_UnpaddedDelete(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)
	rec := encryptUnpadded(t, svc, "gone.bin", []byte("bye"))
	path, _ := encryptedFilePath(baseDir, rec.FileID)

	if _, code := svc.DeleteStored(context.Background(), rec.FileID); code != http.StatusOK {
		t.Fatalf("DeleteStored status = %d, want 200", code)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Error("encrypted file still on disk after delete")
	}
}

// TestFileCrypt_UnpaddedRoundTripAcrossSizes walks the chunk boundary, where
// the final-chunk flag and the tag count are decided.
func TestFileCrypt_UnpaddedRoundTripAcrossSizes(t *testing.T) {
	svc, _, baseDir := newFileCryptService(t)

	for _, size := range []int{1, 8, fileChunkSize - 1, fileChunkSize, fileChunkSize + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			content := randomContent(t, size)
			rec := encryptUnpadded(t, svc, "sized.bin", content)

			if got, want := storedSize(t, baseDir, rec.FileID), plainSealedSize(int64(size)); got != want {
				t.Errorf("sealed size = %d, want %d", got, want)
			}
			_, data, err := download(t, svc, rec.FileID)
			if err != nil {
				t.Fatalf("WriteTo error: %v", err)
			}
			if !bytes.Equal(data, content) {
				t.Errorf("round trip of %d bytes returned %d bytes of different content", size, len(data))
			}
		})
	}
}
