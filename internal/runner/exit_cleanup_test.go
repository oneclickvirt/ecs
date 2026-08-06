package runner

import (
	"sync"
	"testing"
)

func TestRunExitCleanupRunsOnceConcurrently(t *testing.T) {
	var mu sync.Mutex
	calls := 0
	SetExitCleanup(func() {
		mu.Lock()
		calls++
		mu.Unlock()
	})
	defer SetExitCleanup(nil)

	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runExitCleanup()
		}()
	}
	wg.Wait()
	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("cleanup callback called %d times, want once", calls)
	}
}

func TestRunExitCleanupRecoversPanic(t *testing.T) {
	called := false
	SetExitCleanup(func() {
		called = true
		panic("cleanup failure")
	})
	defer SetExitCleanup(nil)
	runExitCleanup()
	if !called {
		t.Fatal("cleanup callback was not called")
	}
	// A recovered callback remains one-shot even after a panic.
	runExitCleanup()
}
