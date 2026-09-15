package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// encryptedFileExt is the extension given to encrypted files on disk.
	encryptedFileExt = ".enc"

	// sealedFilePerm is the mode for sealed files, the same one the envelope
	// package uses.
	sealedFilePerm = 0o600
)

// fileIDPattern allows lowercase hex only. Checking it before building a path
// blocks path traversal: "../../etc/passwd" can never match.
var fileIDPattern = regexp.MustCompile(`^[a-f0-9]{16,64}$`)

// Errors returned by the storage helpers.
var (
	// ErrInvalidFileID means a file id is not valid hex.
	ErrInvalidFileID = errors.New("service: invalid file id")

	// ErrFileTooLarge means an upload is, or claims to be, over the size limit.
	ErrFileTooLarge = errors.New("service: file exceeds the maximum allowed size")

	// ErrEmptyUpload means an upload has no bytes.
	ErrEmptyUpload = errors.New("service: upload is empty")
)

// validFileID reports whether id is a safe, well-formed file id.
func validFileID(id string) bool {
	return fileIDPattern.MatchString(id)
}

// ensureDir creates dir (and any parents) if it does not already exist.
func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0o750)
}

// sanitizeName turns an uploaded filename into a safe display name: base name
// only, no control characters or quotes, trimmed and length-capped. It returns
// "file" if nothing usable is left.
func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	// keep only the last path element, for both "/" and "\" separators
	if i := strings.LastIndexAny(name, `/\`); i >= 0 {
		name = name[i+1:]
	}

	var b strings.Builder
	for _, r := range name {
		switch {
		case r < 0x20 || r == 0x7f: // control characters
			continue
		case r == '"' || r == '\\' || r == '/': // quotes and path separators
			continue
		default:
			b.WriteRune(r)
		}
	}
	cleaned := strings.TrimSpace(b.String())
	cleaned = strings.Trim(cleaned, ".") // avoid "", "." and ".."

	const maxNameLen = 128
	if len(cleaned) > maxNameLen {
		cleaned = cleaned[len(cleaned)-maxNameLen:]
	}
	if cleaned == "" {
		return "file"
	}
	return cleaned
}

// encryptedFilePath returns the path of id's encrypted file inside baseDir. A
// bad id gets ErrInvalidFileID, so the path can never leave baseDir.
func encryptedFilePath(baseDir, id string) (string, error) {
	if !validFileID(id) {
		return "", ErrInvalidFileID
	}
	return filepath.Join(baseDir, id+encryptedFileExt), nil
}

// limitedReader reads from r and fails with ErrFileTooLarge once the source
// goes over its limit.
type limitedReader struct {
	r    io.Reader
	left int64 // bytes still allowed, plus one slack byte
}

// newLimitedReader limits r to limit bytes. It allows one extra byte, so a
// source of exactly limit bytes passes and a longer one is caught as soon as
// that byte arrives.
func newLimitedReader(r io.Reader, limit int64) *limitedReader {
	return &limitedReader{r: r, left: limit + 1}
}

// Read implements io.Reader.
func (l *limitedReader) Read(p []byte) (int, error) {
	if l.left <= 0 {
		return 0, ErrFileTooLarge
	}
	if int64(len(p)) > l.left {
		p = p[:l.left]
	}

	n, err := l.r.Read(p)
	l.left -= int64(n)
	if l.left <= 0 {
		// the slack byte arrived, so the source is over the limit. Decide it on
		// this read: a reader may return io.EOF with its last bytes (multipart
		// parts do), which would otherwise look like a complete file.
		return n, ErrFileTooLarge
	}
	return n, err
}

// sealedFile is what the record needs about a stored file, counted as the bytes
// went by.
type sealedFile struct {
	size int64  // exact payload length
	sum  string // sha-256 of the payload, hex encoded
}

// sealSized seals src into path, padded, in one pass. The client declared size
// up front, so the padding is sized before reading starts. If src does not
// match that size, it fails before any padding is written.
func sealSized(masterKey []byte, path, fileID string, src io.Reader, size, maxSize int64) (sealedFile, error) {
	// a declared length lets both be checked before reading
	if size > maxSize {
		return sealedFile{}, ErrFileTooLarge
	}
	if size == 0 {
		return sealedFile{}, ErrEmptyUpload
	}

	digest := sha256.New()
	if err := sealPaddedTo(masterKey, path, fileID, io.TeeReader(src, digest), size); err != nil {
		return sealedFile{}, err
	}
	return sealedFile{size: size, sum: hex.EncodeToString(digest.Sum(nil))}, nil
}

// sealSizeless seals src into path, padded, with no declared length (a
// multipart part or a chunked body). It is still one pass: the length is
// counted as src streams in, and chunk 0, which holds that length, is written
// last. That costs one extra chunk of memory and needs a file it can write by
// offset.
func sealSizeless(masterKey []byte, path, fileID string, src io.Reader, maxSize int64) (sealedFile, error) {
	return sealCounted(path, src, maxSize, func(dst *os.File, src io.Reader) (int64, error) {
		return scheme.SealPaddedAtAAD(masterKey, dst, src, []byte(fileID))
	})
}

// sealPlain seals src into path without padding, so no length is needed. The
// catch: the file size gives away the plaintext length.
func sealPlain(masterKey []byte, path, fileID string, src io.Reader, maxSize int64) (sealedFile, error) {
	return sealCounted(path, src, maxSize, func(dst *os.File, src io.Reader) (int64, error) {
		return scheme.SealStreamAAD(masterKey, dst, src, []byte(fileID))
	})
}

// sealCounted creates path and runs a seal that takes no length, hashing the
// bytes on the way in. Neither sizeless seal knows the length until src ends,
// so an empty upload is only caught afterwards.
func sealCounted(path string, src io.Reader, maxSize int64, seal func(dst *os.File, src io.Reader) (int64, error)) (sealedFile, error) {
	digest := sha256.New()
	limited := newLimitedReader(io.TeeReader(src, digest), maxSize)

	var n int64
	err := sealTo(path, func(dst *os.File) error {
		var err error
		n, err = seal(dst, limited)
		return err
	})
	if err != nil {
		return sealedFile{}, err
	}

	// an empty upload only shows up once src has ended
	if n == 0 {
		_ = os.Remove(path)
		return sealedFile{}, ErrEmptyUpload
	}
	return sealedFile{size: n, sum: hex.EncodeToString(digest.Sum(nil))}, nil
}

// sealPaddedTo creates path and seals exactly size bytes of src into it, padded
// to a Padmé bucket. Nothing is staged; plaintext only exists in the chunk
// being sealed.
func sealPaddedTo(masterKey []byte, path, fileID string, src io.Reader, size int64) error {
	return sealTo(path, func(dst *os.File) error {
		_, err := scheme.SealPaddedStreamAAD(masterKey, dst, src, size, []byte(fileID))
		return err
	})
}

// sealTo creates path, hands the open file to seal, and syncs it before
// closing. If anything fails, the file is removed.
//
// seal gets an *os.File, not an io.Writer, because the sizeless padded seal
// writes chunk 0 last, by offset.
//
// Every seal binds the file id in as AAD, so a file moved to another id no
// longer opens.
func sealTo(path string, seal func(dst *os.File) error) error {
	// O_EXCL: never overwrite an existing file. encryptedFilePath keeps path
	// inside the encrypted-files directory, so gosec's G304 does not apply.
	dst, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, sealedFilePerm) // #nosec G304
	if err != nil {
		return err
	}

	err = seal(dst)
	if err == nil {
		err = dst.Sync()
	}
	if closeErr := dst.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		// never leave a half-sealed file where a download could find it
		_ = os.Remove(path)
		return err
	}

	return nil
}
