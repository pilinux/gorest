package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pilinux/crypt/envelope"
	gmodel "github.com/pilinux/gorest/database/model"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/pilinux/gorest/example3/internal/database/model"
	"github.com/pilinux/gorest/example3/internal/repo"
)

// fileIDBytes is how many random bytes a file id has. It is hex-encoded, so the
// string is twice as long.
const fileIDBytes = 16

// SizeUnknown tells EncryptAndStore the client sent no length, so it is counted
// while the upload streams in.
const SizeUnknown = -1

// padded and unpadded are values for the record's Unpadded flag, so callers of
// storeSealed can say which format they sealed.
const (
	padded   = false
	unpadded = true
)

// ErrIntegrityCheckFailed means a decrypted file does not match the size
// recorded when it was encrypted.
var ErrIntegrityCheckFailed = errors.New("service: decrypted file failed its integrity check")

// FileCryptService encrypts uploaded files, stores them on disk, and keeps
// their metadata in MongoDB.
//
// Files are sealed a chunk at a time, so memory use stays small however big the
// file is. The payload is padded to a Padmé bucket, so the stored size does not
// give away the exact length. The upload is read only once, and no plaintext
// ever touches the disk.
type FileCryptService struct {
	keys    masterKeyProvider
	store   repo.FileRecordStore
	baseDir string
	maxSize int64 // largest accepted plaintext, in bytes
}

// NewFileCryptService returns a FileCryptService that keeps encrypted files in
// baseDir and rejects uploads larger than maxSize bytes.
func NewFileCryptService(keys masterKeyProvider, store repo.FileRecordStore, baseDir string, maxSize int64) *FileCryptService {
	return &FileCryptService{
		keys:    keys,
		store:   store,
		baseDir: baseDir,
		maxSize: maxSize,
	}
}

// MaxSize returns the largest file this service accepts, in bytes. The handler
// uses it to limit the whole request body.
func (s *FileCryptService) MaxSize() int64 {
	return s.maxSize
}

// EncryptAndStore encrypts src into a padded file under a new file id and saves
// its record. On success the response holds the new FileRecord.
//
// size is the length the client declared, or SizeUnknown. Either way src is
// read once, straight into the stored file; without a declared length, it is
// counted on the way in and chunk 0 is written last. src is never buffered
// whole or written anywhere in the clear.
func (s *FileCryptService) EncryptAndStore(ctx context.Context, name string, src io.Reader, size int64) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	return s.storeSealed(ctx, name, func(masterKey []byte, path, fileID string) (sealedFile, error) {
		if size == SizeUnknown {
			return sealSizeless(masterKey, path, fileID, src, s.maxSize)
		}
		return sealSized(masterKey, path, fileID, src, size, s.maxSize)
	}, padded)
}

// EncryptAndStoreUnpadded is EncryptAndStore without padding. No length is
// needed, so the file is simply written in order. The catch: the file size
// gives away the plaintext length. Use it only when that does not matter.
func (s *FileCryptService) EncryptAndStoreUnpadded(ctx context.Context, name string, src io.Reader) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	return s.storeSealed(ctx, name, func(masterKey []byte, path, fileID string) (sealedFile, error) {
		return sealPlain(masterKey, path, fileID, src, s.maxSize)
	}, unpadded)
}

// storeSealed runs seal under a new file id and saves the record. It holds the
// steps both encrypt methods share: key, id, path, error mapping and rollback.
func (s *FileCryptService) storeSealed(
	ctx context.Context,
	name string,
	seal func(masterKey []byte, path, fileID string) (sealedFile, error),
	unpadded bool,
) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	masterKey, err := s.keys.MasterKey()
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("storeSealed.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	fileID, err := envelope.RandomHex(fileIDBytes)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("storeSealed.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if err := ensureDir(s.baseDir); err != nil {
		log.WithContext(ctx).WithError(err).Error("storeSealed.s.3")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	path, err := encryptedFilePath(s.baseDir, fileID)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("storeSealed.s.4")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	up, err := seal(masterKey, path, fileID)
	if err != nil {
		httpResponse.Message, httpStatusCode = classifyUploadError(err)
		if httpStatusCode == http.StatusInternalServerError {
			log.WithContext(ctx).WithError(err).Error("storeSealed.s.5")
		}
		return
	}

	rec := &model.FileRecord{
		FileID:    fileID,
		Name:      sanitizeName(name),
		Size:      up.size,
		Unpadded:  unpadded,
		CreatedAt: time.Now().Unix(),
	}
	err = sealRecord(masterKey, rec)
	if err == nil {
		err = s.store.Create(ctx, rec)
	}
	if err != nil {
		// delete the file so no ciphertext is left without a record
		_ = os.Remove(path)
		log.WithContext(ctx).WithError(err).Error("storeSealed.s.6")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = rec
	httpStatusCode = http.StatusCreated
	return
}

// sealRecord seals the record's name and size into the fields that get stored.
func sealRecord(masterKey []byte, rec *model.FileRecord) (err error) {
	rec.SealedName, err = scheme.SealStringAAD(masterKey, rec.Name, []byte(nameAADLabel+rec.FileID))
	if err != nil {
		return err
	}
	rec.SealedSize, err = scheme.SealInt64AAD(masterKey, rec.Size, []byte(sizeAADLabel+rec.FileID))
	return err
}

// openRecord fills in the record's name and size from their sealed copies.
func openRecord(masterKey []byte, rec *model.FileRecord) (err error) {
	rec.Name, err = scheme.OpenStringAAD(masterKey, rec.SealedName, []byte(nameAADLabel+rec.FileID))
	if err != nil {
		return err
	}
	rec.Size, err = scheme.OpenInt64AAD(masterKey, rec.SealedSize, []byte(sizeAADLabel+rec.FileID))
	return err
}

// classifyUploadError maps a failed seal to a response. Client mistakes get a
// clear message; anything else is a 500.
func classifyUploadError(err error) (message any, httpStatusCode int) {
	switch {
	case errors.Is(err, ErrEmptyUpload):
		return "empty file", http.StatusBadRequest

	case errors.Is(err, ErrFileTooLarge):
		return "file is too large", http.StatusRequestEntityTooLarge

	// the file fit, but the whole request body went over the handler's limit
	case errors.As(err, new(*http.MaxBytesError)):
		return "request body is too large", http.StatusRequestEntityTooLarge

	// the body did not match its declared length (caught before any padding
	// was written). Too short may be worth a retry; too long was declared wrong.
	case errors.Is(err, envelope.ErrSourceShort):
		return "upload ended before its declared length", http.StatusBadRequest
	case errors.Is(err, envelope.ErrSourceLong):
		return "upload is longer than its declared length", http.StatusBadRequest
	case errors.Is(err, envelope.ErrSourceSize):
		return "upload does not match its declared length", http.StatusBadRequest

	// the body was cut off mid-stream: a dropped connection, or a chunked or
	// multipart body that ended early. The client's fault, not the server's.
	case errors.Is(err, io.ErrUnexpectedEOF):
		return "upload ended unexpectedly", http.StatusBadRequest

	default:
		return "internal server error", http.StatusInternalServerError
	}
}

// Download is an open plaintext stream of one stored file. Everything that can
// be checked up front already has been, so the caller can send the status and
// length before calling WriteTo. Close it when done.
type Download struct {
	Name string // sanitized display name, for the download header
	Size int64  // plaintext size, as recorded at encrypt time

	file   *os.File               // the encrypted file on disk
	padded *envelope.PaddedReader // payload of a padded file; nil for an unpadded one
	plain  *envelope.StreamReader // plaintext of an unpadded file; nil for a padded one
}

// OpenForDownload finds fileID, opens its encrypted file, and returns a
// Download ready to stream.
//
// For a padded file, the first chunk is authenticated here, so a wrong key, a
// bad file or a size mismatch fails now with a 500 instead of cutting off a
// 200. An unpadded file only has its header read here; its first chunk is
// checked while the body is being written.
func (s *FileCryptService) OpenForDownload(ctx context.Context, fileID string) (dl *Download, httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if !validFileID(fileID) {
		httpResponse.Message = "invalid file id"
		httpStatusCode = http.StatusBadRequest
		return
	}

	rec, err := s.store.GetByFileID(ctx, fileID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "file not found"
			httpStatusCode = http.StatusNotFound
			return
		}
		log.WithContext(ctx).WithError(err).Error("OpenForDownload.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	dl, err = s.openSealed(rec)
	if err != nil {
		// metadata exists but the file is gone from disk
		if errors.Is(err, os.ErrNotExist) {
			httpResponse.Message = "file not found"
			httpStatusCode = http.StatusNotFound
			return
		}
		log.WithContext(ctx).WithError(err).Error("OpenForDownload.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpStatusCode = http.StatusOK
	return
}

// openSealed opens rec's encrypted file with the reader its format needs. It
// returns a raw error so the caller can tell failures apart.
//
// Both formats use the file id as AAD, plus a tag of their own, so neither
// reader accepts the other's file. A record with a flipped flag fails to open
// instead of being misread.
func (s *FileCryptService) openSealed(rec *model.FileRecord) (*Download, error) {
	masterKey, err := s.keys.MasterKey()
	if err != nil {
		return nil, err
	}
	if err := openRecord(masterKey, rec); err != nil {
		return nil, err
	}

	path, err := encryptedFilePath(s.baseDir, rec.FileID)
	if err != nil {
		return nil, err
	}
	// safe: encryptedFilePath checks the id is hex and keeps the path inside
	// baseDir
	f, err := os.Open(path) // #nosec G304
	if err != nil {
		return nil, err
	}

	dl := &Download{
		Name: rec.Name,
		Size: rec.Size,
		file: f,
	}
	aad := []byte(rec.FileID)
	if rec.Unpadded {
		// reads only the header; nothing is authenticated until the first chunk
		dl.plain, err = scheme.OpenReaderAAD(masterKey, f, aad)
	} else {
		// authenticates the first chunk to read the sealed length. A record
		// claiming a different length is refused before anything is sent.
		dl.padded, err = scheme.OpenPaddedReaderAAD(masterKey, f, aad)
		if err == nil && dl.padded.Size() != rec.Size {
			// closing wipes the chunk decrypted to read the length
			_ = dl.padded.Close()
			err = ErrIntegrityCheckFailed
		}
	}
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return dl, nil
}

// WriteTo streams the plaintext to dst and returns how many bytes it produced.
//
// It writes no second file on disk. Each chunk is authenticated before it
// goes out, but a damaged file can still fail midway, so treat dst as bad
// unless the error is nil. A nil error means the whole stream checked out,
// padding included, and the size matches the record.
func (d *Download) WriteTo(dst io.Writer) (int64, error) {
	if d.padded != nil {
		// writes only the payload, but still reads and checks the padding after
		// it, so a stream cut off inside its padding does not pass
		return d.padded.WriteTo(dst)
	}
	return d.writePlain(dst)
}

// writePlain copies an unpadded file's plaintext to dst. A plain stream has no
// sealed length, so it must match the recorded size exactly: no shorter, no
// longer.
func (d *Download) writePlain(dst io.Writer) (int64, error) {
	n, err := io.CopyN(dst, d.plain, d.Size)
	if err != nil {
		if errors.Is(err, io.EOF) {
			// the stream ended properly, but shorter than the record says
			return n, ErrIntegrityCheckFailed
		}
		return n, err
	}

	// read one byte, not the rest: a record claiming too little should not
	// force reading the whole file. io.EOF only comes after an authentic final
	// chunk, so a cut-off stream fails here too.
	var probe [1]byte
	switch m, err := io.ReadFull(d.plain, probe[:]); {
	case m > 0:
		return n, ErrIntegrityCheckFailed
	case errors.Is(err, io.EOF):
		return n, nil
	default:
		return n, err
	}
}

// Close ends the plaintext stream and closes the encrypted file.
//
// If a download stopped midway, the reader may still hold a decrypted chunk;
// Close wipes it. It does not re-check the stream: WriteTo's result stands.
func (d *Download) Close() error {
	if d.padded != nil {
		// may return ErrIncompleteRead, which WriteTo already reported; only the
		// wipe matters here
		_ = d.padded.Close()
	}
	if d.plain != nil {
		d.plain.Abort()
	}
	return d.file.Close()
}

// DeleteStored deletes the encrypted file and the record for fileID.
func (s *FileCryptService) DeleteStored(ctx context.Context, fileID string) (httpResponse gmodel.HTTPResponse, httpStatusCode int) {
	if !validFileID(fileID) {
		httpResponse.Message = "invalid file id"
		httpStatusCode = http.StatusBadRequest
		return
	}

	path, err := encryptedFilePath(s.baseDir, fileID)
	if err != nil {
		log.WithContext(ctx).WithError(err).Error("DeleteStored.s.1")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	// delete the file before the record, so a failure at either step leaves the
	// record in place and a retry can finish the job; if the file is already
	// gone, that is fine
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.WithContext(ctx).WithError(err).Error("DeleteStored.s.2")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	if err := s.store.DeleteByFileID(ctx, fileID); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			httpResponse.Message = "file not found"
			httpStatusCode = http.StatusNotFound
			return
		}
		log.WithContext(ctx).WithError(err).Error("DeleteStored.s.3")
		httpResponse.Message = "internal server error"
		httpStatusCode = http.StatusInternalServerError
		return
	}

	httpResponse.Message = "file deleted successfully"
	httpStatusCode = http.StatusOK
	return
}
