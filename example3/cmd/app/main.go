// Package main is the entry point of the example3 application: a MongoDB-only
// demo of envelope encryption with a rotatable env secret.
package main

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	gconfig "github.com/pilinux/gorest/config"
	gdb "github.com/pilinux/gorest/database"
	gserver "github.com/pilinux/gorest/lib/server"

	"github.com/pilinux/gorest/example3/internal/repo"
	"github.com/pilinux/gorest/example3/internal/router"
	"github.com/pilinux/gorest/example3/internal/service"
)

const (
	// defaultEncryptedFilesDir is where encrypted files are written when
	// ENCRYPTED_FILES_DIR is not set.
	defaultEncryptedFilesDir = "encrypted-files"

	// defaultMaxUploadSize is the largest accepted upload (bytes) when
	// MAX_UPLOAD_SIZE_MB is not set: 25 MiB.
	defaultMaxUploadSize = 25 << 20
)

func main() {
	// load configuration from the environment
	if err := gconfig.Config(); err != nil {
		fmt.Println(err)
		return
	}
	configure := gconfig.GetConfig()

	// example3 is MongoDB-only: the wrapped master key and the file metadata
	// both live in Mongo.
	if !gconfig.IsMongo() {
		fmt.Println("ACTIVATE_MONGO must be set to yes for example3")
		return
	}
	for {
		if _, err := gdb.InitMongo(); err != nil {
			fmt.Println(err)
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	// create the indexes before the master key is touched: the unique index on
	// keyName is what stops two instances booting at the same time from each
	// storing a master key of their own, and that race is decided on the very
	// first boot.
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := repo.EnsureIndexes(ctx, gdb.GetMongo())
		cancel()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	// derive the KEK(s) from the env secret(s) and load (or bootstrap/rotate)
	// the master key stored in MongoDB.
	keyManager := service.NewKeyManager()
	if err := keyManager.SetSecrets(
		strings.TrimSpace(os.Getenv("ENCRYPTION_SECRET")),
		strings.TrimSpace(os.Getenv("ENCRYPTION_SECRET_OLD")),
	); err != nil {
		fmt.Println(err)
		return
	}
	{
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := keyManager.EnsureAndLoad(ctx, repo.NewKeyStoreRepo(gdb.GetMongo()))
		cancel()
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	encryptedFilesDir := strings.TrimSpace(os.Getenv("ENCRYPTED_FILES_DIR"))
	if encryptedFilesDir == "" {
		encryptedFilesDir = defaultEncryptedFilesDir
	}

	r, err := router.SetupRouter(configure, keyManager, encryptedFilesDir, maxUploadSize())
	if err != nil {
		fmt.Println(err)
		return
	}

	// attach the router to a http.Server with timeouts. The read budget is
	// short; file uploads set their own, longer one. Files are streamed out
	// too, and an upload answers only after its transfer, so the write budget
	// has to cover a whole transfer.
	srv := &http.Server{
		Addr:              configure.Server.ServerHost + ":" + configure.Server.ServerPort,
		Handler:           r,
		ReadTimeout:       1 * time.Minute,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	// graceful shutdown
	const shutdownTimeout = 30 * time.Second
	done := make(chan struct{})
	go func() {
		if err := gserver.GracefulShutdown(srv, shutdownTimeout, done, gdb.CloseAllDB); err != nil {
			fmt.Println(err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("server error: %v\n", err)
	}
	<-done
	fmt.Println("server shutdown complete")
}

// maxUploadSize reads MAX_UPLOAD_SIZE_MB (in MiB) and returns it in bytes,
// falling back to defaultMaxUploadSize.
func maxUploadSize() int64 {
	raw := strings.TrimSpace(os.Getenv("MAX_UPLOAD_SIZE_MB"))
	if raw == "" {
		return defaultMaxUploadSize
	}
	mb, err := strconv.ParseInt(raw, 10, 64)
	// too large a value would overflow the shift into a negative limit
	if err != nil || mb <= 0 || mb > math.MaxInt64>>20 {
		return defaultMaxUploadSize
	}
	return mb << 20
}
