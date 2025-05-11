// Package server provides utilities for managing HTTP server lifecycle operations.
package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulShutdown listens for OS signals and gracefully shuts down the server
func GracefulShutdown(srv *http.Server, timeout time.Duration, done chan struct{}, closeDB ...func() error) (err error) {
	// create context that listens for the interrupt signal from the OS
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	log.Printf("signal %v received: starting graceful shutdown with timeout %v\n", sig, timeout)

	// the context is used to inform the server that it has 'x' seconds
	// to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// shutdown the server in goroutine with a timeout
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Shutdown(ctx)
	}()
	select {
	case err = <-errCh:
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				// shutdown returned a timeout error
				log.Printf("graceful shutdown deadline exceeded, forcefully closing server: %v\n", err)
				if eThis := srv.Close(); eThis != nil {
					// forcefully closing the server also failed, log the error
					err = fmt.Errorf("server force-close error 1: %w", eThis)
				}
			} else {
				// shutdown returned any other error
				err = fmt.Errorf("graceful shutdown error: %w", err)
			}
		}
	case <-ctx.Done():
		// context deadline exceeded before the server could shutdown gracefully
		log.Printf("graceful shutdown timed out, forcefully closing server: %v\n", err)
		if eThis := srv.Close(); eThis != nil {
			// forcefully closing the server also failed, log the error
			err = fmt.Errorf("server force-close error 2: %w", eThis)
		}
	}

	// shutdown all DB connections
	if len(closeDB) > 0 {
		for _, closeFn := range closeDB {
			if closeFn == nil {
				continue
			}
			if eThis := closeFn(); eThis != nil && err == nil {
				err = fmt.Errorf("error closing DB connections: %w", eThis)
			}
		}
	}

	// notify the main goroutine that the shutdown is complete
	close(done)

	return err
}
