package database_test

import (
	"sync"
	"testing"

	gdb "github.com/pilinux/gorest/database"
)

// TestCloseSQL_Concurrent ensures CloseSQL is safe to call concurrently.
func TestCloseSQL_Concurrent(t *testing.T) {
	t.Log("running:" + t.Name())
	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gdb.CloseSQL()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CloseSQL returned error: %v", err)
		}
	}
}

// TestCloseRedis_Concurrent ensures CloseRedis is safe to call concurrently.
func TestCloseRedis_Concurrent(t *testing.T) {
	t.Log("running:" + t.Name())
	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gdb.CloseRedis()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CloseRedis returned error: %v", err)
		}
	}
}

// TestCloseMongo_Concurrent ensures CloseMongo is safe to call concurrently.
func TestCloseMongo_Concurrent(t *testing.T) {
	t.Log("running:" + t.Name())
	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gdb.CloseMongo()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CloseMongo returned error: %v", err)
		}
	}
}

// TestCloseAllDB_Concurrent ensures CloseAllDB is safe to call concurrently.
func TestCloseAllDB_Concurrent(t *testing.T) {
	t.Log("running:" + t.Name())

	const goroutines = 50
	var wg sync.WaitGroup
	errs := make(chan error, goroutines)

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs <- gdb.CloseAllDB()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("CloseAllDB returned error: %v", err)
		}
	}
}
