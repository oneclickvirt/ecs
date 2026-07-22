//go:build !ecs_public

package tests

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/oneclickvirt/privatespeedtest/pst"
)

func resetPrivateSpeedRegistryForTest(t *testing.T, loader func() (*pst.ServerList, error)) {
	t.Helper()
	originalLoader := privateSpeedRegistryLoader
	privateSpeedRegistryOnce = sync.Once{}
	privateSpeedRegistry = nil
	privateSpeedRegistryErr = nil
	privateSpeedRegistryLoader = loader
	t.Cleanup(func() {
		privateSpeedRegistryOnce = sync.Once{}
		privateSpeedRegistry = nil
		privateSpeedRegistryErr = nil
		privateSpeedRegistryLoader = originalLoader
	})
}

func TestPrivateSpeedServerListLoadsOnceAcrossCarrierGroups(t *testing.T) {
	var calls atomic.Int32
	want := &pst.ServerList{TotalServers: 1}
	resetPrivateSpeedRegistryForTest(t, func() (*pst.ServerList, error) {
		calls.Add(1)
		return want, nil
	})

	const callers = 8
	var wg sync.WaitGroup
	wg.Add(callers)
	for range callers {
		go func() {
			defer wg.Done()
			got, err := privateSpeedServerList()
			if err != nil || got != want {
				t.Errorf("privateSpeedServerList() = %#v, %v", got, err)
			}
		}()
	}
	wg.Wait()
	if got := calls.Load(); got != 1 {
		t.Fatalf("registry loader calls = %d, want 1", got)
	}
}

func TestPrivateSpeedServerListCachesStableFailure(t *testing.T) {
	var calls atomic.Int32
	wantErr := errors.New("registry unavailable")
	resetPrivateSpeedRegistryForTest(t, func() (*pst.ServerList, error) {
		calls.Add(1)
		return nil, wantErr
	})

	for range 3 {
		got, err := privateSpeedServerList()
		if got != nil || !errors.Is(err, wantErr) {
			t.Fatalf("privateSpeedServerList() = %#v, %v", got, err)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("registry loader calls = %d, want 1", got)
	}
}

func TestLoadPrivateSpeedRegistryHasValidatedFallback(t *testing.T) {
	loaded, err := pst.LoadEmbeddedServerList()
	if err != nil {
		t.Fatal(err)
	}
	if loaded.List == nil || loaded.Metadata.Count < 10 || loaded.Source != "embedded" || !loaded.Fallback {
		t.Fatalf("unexpected embedded fallback: %#v", loaded)
	}
}
