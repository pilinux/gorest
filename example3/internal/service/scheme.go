package service

import "github.com/pilinux/crypt/envelope"

// HKDF labels for example3. Never change them: data wrapped or sealed under
// these labels can only be opened with the same ones.
const (
	kekLabel    = "gorest/example3:kek:v1"
	subKeyLabel = "gorest/example3:data-subkey:v1"
)

// AAD labels for a file record's sealed name and size. Never change them
// either. Each field gets its own, so the two can't be swapped.
const (
	nameAADLabel = "gorest/example3:file-name:v1:"
	sizeAADLabel = "gorest/example3:file-size:v1:"
)

// AAD labels for text and number tokens. Never change them either. Each type
// gets its own, so a number token can't be opened as text, or the other way.
const (
	textAADLabel  = "gorest/example3:text:v1"
	int64AADLabel = "gorest/example3:int64:v1"
)

// fileChunkSize is how much plaintext is sealed at a time. A request holds a
// chunk or two in memory, and each chunk adds a 16-byte tag. Unlike the labels,
// it can change: every stored file records its own chunk size, so old files
// still open as long as it stays within maxAcceptedChunkSize.
const fileChunkSize = envelope.DefaultChunkSize // 1 MiB

// maxAcceptedChunkSize is the largest chunk size a reader accepts from a file
// header. The chunk buffer is allocated before anything is authenticated, so
// without this limit a tiny forged file could make each download take 64 MiB.
//
// It is kept separate from fileChunkSize because it must stay at or above every
// chunk size a stored file was written with. Raising fileChunkSize past it
// makes the first seal fail with envelope.ErrInvalidChunkSize, so the two
// cannot drift apart unnoticed.
const maxAcceptedChunkSize = 1 << 20 // 1 MiB

// scheme is the envelope scheme all example3 services share. It never changes
// after setup and is safe for concurrent use.
//
// The envelope package binds a format tag into every chunk, so padded and plain
// readers never accept each other's files, whatever AAD is passed.
var scheme = envelope.New(envelope.Config{
	KEKLabel:             kekLabel,
	SubKeyLabel:          subKeyLabel,
	ChunkSize:            fileChunkSize,
	MaxAcceptedChunkSize: maxAcceptedChunkSize,
})
