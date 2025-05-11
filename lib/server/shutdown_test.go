package server_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/pilinux/gorest/lib/server"
)

// newTestServer sets up a new HTTP server on a random port with a given handler
func newTestServer(t *testing.T, handler http.HandlerFunc) (*http.Server, net.Listener) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on a port: %v", err)
	}

	srv := &http.Server{
		Handler: handler,
		Addr:    ln.Addr().String(),
	}

	go func() {
		if errSrv := srv.Serve(ln); errSrv != nil {
			// ignore "server closed" and "use of closed network connection" errors
			if errors.Is(errSrv, http.ErrServerClosed) || errors.Is(errSrv, net.ErrClosed) {
				return
			}
			t.Errorf("failed to serve: %v", errSrv)
		}
	}()
	time.Sleep(50 * time.Millisecond) // wait for server to start
	return srv, ln
}

// sendSignal sends the specified OS signal to the current process after a delay
func sendSignal(sig os.Signal, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		if s, ok := sig.(syscall.Signal); ok {
			if err := syscall.Kill(syscall.Getpid(), s); err != nil {
				log.Printf("failed to send signal %v: %v", sig, err)
			}
		} else {
			log.Printf("failed to assert signal %v to syscall.Signal", sig)
		}
	}()
}

// createHangingHandler creates an HTTP handler that sleeps for a specified duration
func createHangingHandler(sleepDuration time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(sleepDuration)
		_, _ = fmt.Fprintln(w, "finally responded")
	}
}

// TestGracefulShutdown_SuccessWithDB tests that the server shuts down gracefully
// within the timeout, with successful closeDB functions including nil ones.
func TestGracefulShutdown_SuccessWithDB(t *testing.T) {
	t.Log("running:" + t.Name())

	// set up a server with a quick handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "Hello, World!")
		if err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})
	srv, ln := newTestServer(t, handler)
	defer func() {
		_ = ln.Close()
	}()

	// define closeDB functions with call tracking
	called1 := false
	closeDB1 := func() error {
		called1 = true
		return nil
	}
	called2 := false
	closeDB2 := func() error {
		called2 = true
		return nil
	}

	// channel to signal shutdown completion
	done := make(chan struct{})

	// run GracefulShutdown with closeDB1, nil, and closeDB2
	go func() {
		err := server.GracefulShutdown(srv, 5*time.Second, done, closeDB1, nil, closeDB2)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	}()

	// send SIGINT after 1 second
	sendSignal(syscall.SIGINT, 1*time.Second)

	// wait for shutdown to complete
	select {
	case <-done:
		// shutdown completed
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown did not complete within 10 seconds")
	}

	// verify that non-nil closeDB functions were called
	if !called1 {
		t.Error("expected closeDB1 to be called")
	}
	if !called2 {
		t.Error("expected closeDB2 to be called")
	}

	// verify server is shut down
	resp, err := http.Get("http://" + srv.Addr)
	if err == nil {
		_ = resp.Body.Close()
		t.Error("expected error when connecting to shut down server")
	}
}

// TestGracefulShutdown_SuccessNoDB tests that the server shuts down gracefully
// when a signal is received, within the timeout, with no closeDB functions.
// This test is fully covered in TestGracefulShutdown_SuccessWithDB.

// TestGracefulShutdown_SuccessWithDBFail tests that the server shuts down gracefully
// within the timeout but returns an error if a closeDB function fails.
func TestGracefulShutdown_SuccessWithDBFail(t *testing.T) {
	t.Log("running:" + t.Name())

	// set up a server with a quick handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, "Hello, World!")
		if err != nil {
			t.Errorf("failed to write response: %v", err)
		}
	})
	srv, ln := newTestServer(t, handler)
	defer func() {
		_ = ln.Close()
	}()

	// define a closeDB function that fails
	expectedErr := errors.New("db close error")
	closeDB := func() error {
		return expectedErr
	}

	// channel to signal shutdown completion
	done := make(chan struct{})
	var shutdownErr error

	// run GracefulShutdown with the failing closeDB function
	go func() {
		shutdownErr = server.GracefulShutdown(srv, 5*time.Second, done, closeDB)
	}()

	// send SIGINT after 1 second
	sendSignal(syscall.SIGINT, 1*time.Second)

	// Wait for shutdown to complete
	select {
	case <-done:
		// shutdown completed
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown did not complete within 10 seconds")
	}

	// verify that the closeDB error is returned
	if shutdownErr == nil {
		t.Error("expected an error, got nil")
	} else if !errors.Is(shutdownErr, expectedErr) {
		t.Errorf("expected error %v to be wrapped in %v", expectedErr, shutdownErr)
	}

	// verify server is shut down
	resp, err := http.Get("http://" + srv.Addr)
	if err == nil {
		_ = resp.Body.Close()
		t.Error("expected error when connecting to shut down server")
	}
}

// TestGracefulShutdown_ShutdownAndDBFail tests that if shutdown fails with an error
// and closeDB also fails, the shutdown error is returned.
// This test is partially covered in TestGracefulShutdown_SuccessWithDBFail.
// To test server shutdown error precedence, we need to mock srv.Shutdown to return a non-timeout error.
// This is not implementable here as it would require a custom http.Server implementation.

// TestGracefulShutdown_TimeoutForceCloseSuccess tests that if shutdown times out,
// srv.Shutdown returns context.DeadlineExceeded, and the server is forcefully closed successfully.
func TestGracefulShutdown_TimeoutForceCloseSuccess(t *testing.T) {
	t.Log("running:" + t.Name())

	// set up a server with a handler that hangs longer than the timeout
	sleepDuration := 5 * time.Second
	handler := createHangingHandler(sleepDuration)
	srv, ln := newTestServer(t, handler)
	defer func() {
		_ = ln.Close()
	}()

	// start a request to keep the server busy
	go func() {
		resp, err := http.Get("http://" + srv.Addr)
		if err == nil {
			_ = resp.Body.Close()
			t.Error("expected error when connecting to hanging server")
		}
	}()

	// wait briefly to ensure the request starts
	time.Sleep(500 * time.Millisecond)

	// channel to signal shutdown completion
	done := make(chan struct{})
	var shutdownErr error

	// run GracefulShutdown with a 2-second timeout
	go func() {
		shutdownErr = server.GracefulShutdown(srv, 2*time.Second, done)
	}()

	// send SIGINT after 1 second, while the handler is still sleeping
	sendSignal(syscall.SIGINT, 1*time.Second)

	// wait for shutdown to complete
	select {
	case <-done:
		// shutdown completed
	case <-time.After(10 * time.Second):
		t.Fatal("shutdown did not complete within 10 seconds")
	}

	// verify that the error is context.DeadlineExceeded
	if !errors.Is(shutdownErr, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", shutdownErr)
	}

	// verify server is shut down
	resp, err := http.Get("http://" + srv.Addr)
	if err == nil {
		_ = resp.Body.Close()
		t.Error("expected error when connecting to shut down server")
	}
}
