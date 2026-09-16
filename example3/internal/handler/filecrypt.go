package handler

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	grenderer "github.com/pilinux/gorest/lib/renderer"
	log "github.com/sirupsen/logrus"

	"github.com/pilinux/gorest/example3/internal/service"
)

const (
	// fileCryptTimeout bounds how long a single file crypto request may run.
	// It has to cover the whole transfer including encryption.
	fileCryptTimeout = 5 * time.Minute

	// fileFormField is the multipart form field carrying the upload.
	fileFormField = "file"

	// multipartContentType and rawContentType are the two upload shapes the
	// encrypt endpoint accepts: a form post, or the file as the whole body.
	multipartContentType = "multipart/form-data"
	rawContentType       = "application/octet-stream"

	// maxMultipartParts bounds how many parts are inspected while looking for
	// the file field, so a client cannot keep the server reading parts forever.
	maxMultipartParts = 16

	// multipartOverhead is how much more than the upload limit a whole
	// multipart body may weigh: part headers, boundaries, and any ordinary form
	// fields sent alongside the file.
	multipartOverhead = 1 << 20 // 1 MiB
)

// errNoFilePart is returned when the multipart body carries no usable file.
var errNoFilePart = errors.New("handler: no file part in the multipart body")

// FileCryptAPI exposes the file encryption endpoints.
type FileCryptAPI struct {
	svc *service.FileCryptService
}

// NewFileCryptAPI returns a new FileCryptAPI. The upload size limit belongs to
// the service, which checks it as the file streams in. The handler just asks
// the service for that limit, so it knows how big a request body to allow.
func NewFileCryptAPI(svc *service.FileCryptService) *FileCryptAPI {
	return &FileCryptAPI{svc: svc}
}

// EncryptFile encrypts an uploaded file, padded, and stores the ciphertext on
// disk.
//
// Endpoint: POST /api/v1/crypto/files/encrypt
//
// Two ways to send the file. Both are read once, straight into the stored file;
// they differ only in when the padding learns how long the file is:
//
//   - multipart/form-data (field "file"): what a form posts. A part does not
//     say how long it is, so the length is counted as the upload streams in.
//     Chunk 0, which records that length, is held back in memory and written
//     last, once the upload has ended. That costs one extra chunk of memory,
//     not a second pass.
//   - application/octet-stream: the file as the whole body, its length in
//     Content-Length. The padding is sized from that before reading starts, so
//     the file is written in order. Name it with a "name" query parameter or a
//     Content-Disposition header.
//
// Either way the body is read as a stream: gin's form parsing is bypassed, so
// the upload is never buffered in memory, never reaches the OS temp directory,
// and never lands on disk in the clear.
//
// Response: 201 with the file metadata (its id is used to decrypt/delete later).
func (api *FileCryptAPI) EncryptFile(c *gin.Context) {
	api.encrypt(c, true)
}

// EncryptFileUnpadded encrypts an uploaded file without padding.
//
// Endpoint: POST /api/v1/crypto/files/encrypt/unpadded
//
// The same two ways to send the file, but without padding, neither one needs
// to know the length up front:
//
//   - multipart/form-data (field "file"): a plain stream is sealed as it
//     arrives and written in order. Nothing is held back, so a form post
//     takes one chunk of memory here instead of the two the padded endpoint
//     needs.
//   - application/octet-stream: the file as the whole body. Content-Length is
//     still checked so an oversized upload is refused before it is read, but a
//     chunked body with no length works too. Name it with a "name" query
//     parameter or a Content-Disposition header.
//
// The trade-off: the size of the stored file gives away the length of the
// original, which is exactly what the padded endpoint hides. Use this one only
// when that does not matter.
//
// Downloads and deletes use the same routes either way, because the record
// remembers which format its file was stored in.
//
// Response: 201 with the file metadata (its id is used to decrypt/delete later).
func (api *FileCryptAPI) EncryptFileUnpadded(c *gin.Context) {
	api.encrypt(c, false)
}

// encrypt reads the file out of whichever body shape the client used and hands
// it to the service.
func (api *FileCryptAPI) encrypt(c *gin.Context, padded bool) {
	// gin's ContentType() has already dropped the boundary parameter
	switch c.ContentType() {
	case multipartContentType:
		api.encryptMultipart(c, padded)
	case rawContentType:
		api.encryptRaw(c, padded)
	default:
		grenderer.Render(c, gin.H{"message": "expected a multipart/form-data or application/octet-stream upload"}, http.StatusUnsupportedMediaType)
	}
}

// encryptRaw seals a body that is nothing but the file. Content-Length gives the
// padding its length up front, so the file is sealed in order.
func (api *FileCryptAPI) encryptRaw(c *gin.Context, padded bool) {
	// worth reading even when the padding does not need it: too big is refused
	// here, unread
	size := c.Request.ContentLength
	if size > api.svc.MaxSize() {
		grenderer.Render(c, gin.H{"message": "file is too large"}, http.StatusRequestEntityTooLarge)
		return
	}
	// a chunked body has no length (-1), so, like a part, its length is
	// counted as it streams in
	if size < 0 {
		size = service.SizeUnknown
	}

	// the file is the whole body, so the file limit is the body limit. Still
	// enforced: Content-Length is only a claim until the bytes turn up.
	body := http.MaxBytesReader(c.Writer, c.Request.Body, api.svc.MaxSize())

	api.store(c, rawUploadName(c), body, size, padded)
}

// rawUploadName is the display name for a raw upload: the "name" query
// parameter, else a Content-Disposition filename. The service sanitizes it and
// falls back to "file".
func rawUploadName(c *gin.Context) string {
	if name := c.Query("name"); name != "" {
		return name
	}
	_, params, err := mime.ParseMediaType(c.GetHeader("Content-Disposition"))
	if err != nil {
		return ""
	}
	return params["filename"]
}

// encryptMultipart seals the file part of a form post. A part's length is not
// known until it ends, so a padded seal counts it on the way in and writes
// chunk 0 last.
func (api *FileCryptAPI) encryptMultipart(c *gin.Context, padded bool) {
	// limit the whole body, not just the file. The service already limits the
	// file as it streams in, but to find the file, NextPart has to read through
	// every part that comes before it. Without this limit, a client could keep
	// sending form fields and hold the server busy until its read timeout.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, api.svc.MaxSize()+multipartOverhead)

	mr, err := c.Request.MultipartReader()
	if err != nil {
		grenderer.Render(c, gin.H{"message": "expected a multipart/form-data upload"}, http.StatusBadRequest)
		return
	}

	part, err := filePart(mr, fileFormField)
	if err != nil {
		// the cap above ran out before a file part turned up
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			grenderer.Render(c, gin.H{"message": "request body is too large"}, http.StatusRequestEntityTooLarge)
			return
		}
		grenderer.Render(c, gin.H{"message": "no file uploaded"}, http.StatusBadRequest)
		return
	}
	defer func() {
		_ = part.Close()
	}()

	// no length to declare: RFC 7578 asks for no per-part Content-Length and
	// browsers send none
	api.store(c, part.FileName(), part, service.SizeUnknown, padded)
}

// store hands the upload to the padded or the plain seal and renders the
// answer. size is of interest only to the padded one.
func (api *FileCryptAPI) store(c *gin.Context, name string, src io.Reader, size int64, padded bool) {
	// the timeout has to cover the whole transfer: the upload is sealed while
	// it arrives, so a slow client holds the request open for as long as it
	// takes to stream the file in
	ctx, cancel := context.WithTimeout(c.Request.Context(), fileCryptTimeout)
	defer cancel()

	if padded {
		resp, statusCode := api.svc.EncryptAndStore(ctx, name, src, size)
		grenderer.Render(c, resp, statusCode)
		return
	}
	resp, statusCode := api.svc.EncryptAndStoreUnpadded(ctx, name, src)
	grenderer.Render(c, resp, statusCode)
}

// filePart looks through the multipart body and returns the first part that
// holds a file in the given form field. Any other form fields sent along with
// the file are skipped. The part is returned still open, and closing it is up
// to the caller.
func filePart(mr *multipart.Reader, field string) (*multipart.Part, error) {
	for range maxMultipartParts {
		part, err := mr.NextPart()
		if err != nil {
			// io.EOF here means the field was never sent
			return nil, err
		}
		if part.FormName() == field && part.FileName() != "" {
			return part, nil
		}
		_ = part.Close()
	}
	return nil, errNoFilePart
}

// DecryptFile decrypts a stored file and streams the original bytes back.
//
// Endpoint: GET /api/v1/crypto/files/:id/decrypt
//
// The plaintext is written out chunk by chunk as each chunk authenticates, so
// it never exists as a whole anywhere but at the client.
func (api *FileCryptAPI) DecryptFile(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), fileCryptTimeout)
	defer cancel()

	dl, resp, statusCode := api.svc.OpenForDownload(ctx, id)
	if statusCode != http.StatusOK {
		grenderer.Render(c, resp, statusCode)
		return
	}
	defer func() {
		_ = dl.Close()
	}()

	// by now the file is open. For a padded file, its first chunk has also been
	// authenticated and its stored length checked against the record. So the
	// status and the length can be sent now, before any of the body is written.
	c.Header("Content-Disposition", `attachment; filename="`+dl.Name+`"`)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(dl.Size, 10))
	c.Status(http.StatusOK)

	if _, err := dl.WriteTo(c.Writer); err != nil {
		// something went wrong after the headers were already sent: a chunk
		// failed to authenticate, or the client disconnected. It is too late to
		// change the status code, so just stop writing. The body then ends up
		// shorter than the Content-Length promised, and that tells the client
		// the download is incomplete and should not be kept.
		log.WithContext(ctx).WithError(err).Error("DecryptFile.h.1")
	}
}

// DeleteFile removes a stored encrypted file and its metadata.
//
// Endpoint: DELETE /api/v1/crypto/files/:id
func (api *FileCryptAPI) DeleteFile(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))

	ctx, cancel := context.WithTimeout(c.Request.Context(), fileCryptTimeout)
	defer cancel()

	resp, statusCode := api.svc.DeleteStored(ctx, id)
	grenderer.Render(c, resp, statusCode)
}
