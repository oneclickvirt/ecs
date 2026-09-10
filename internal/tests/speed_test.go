//go:build !ecs_public

package tests

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/oneclickvirt/privatespeedtest/pst"
)

func TestPrintTableRowToAlignsPrivateLocationWithSharedTable(t *testing.T) {
	var output bytes.Buffer
	printTableRowTo(&output, pst.SpeedTestResult{
		City:         "Beijing",
		CarrierType:  "Unicom",
		UploadMbps:   1.25,
		DownloadMbps: 2.5,
		PingLatency:  3 * time.Millisecond,
	})

	want := " 联通Beijing     1.25 Mbps       2.50 Mbps       3.00 ms         N/A             \n"
	if got := output.String(); got != want {
		t.Fatalf("private speedtest row = %q, want %q", got, want)
	}
}

func resetPrivateSpeedRegistryForTest(t *testing.T, loader func(pst.Network) (*pst.ServerList, error)) {
	t.Helper()
	privateSpeedRegistryMu.Lock()
	defer privateSpeedRegistryMu.Unlock()
	originalLoader := privateSpeedRegistryLoader
	privateSpeedRegistries = make(map[pst.Network]*privateSpeedRegistryEntry)
	privateSpeedRegistryLoader = loader
	t.Cleanup(func() {
		privateSpeedRegistryMu.Lock()
		defer privateSpeedRegistryMu.Unlock()
		privateSpeedRegistries = make(map[pst.Network]*privateSpeedRegistryEntry)
		privateSpeedRegistryLoader = originalLoader
	})
}

func TestPrivateSpeedServerListLoadsOnceAcrossCarrierGroups(t *testing.T) {
	var calls atomic.Int32
	want := &pst.ServerList{TotalServers: 1}
	resetPrivateSpeedRegistryForTest(t, func(network pst.Network) (*pst.ServerList, error) {
		if network != pst.NetworkIPv4 {
			t.Errorf("registry network = %q, want %q", network, pst.NetworkIPv4)
		}
		calls.Add(1)
		return want, nil
	})

	const callers = 8
	var wg sync.WaitGroup
	wg.Add(callers)
	for range callers {
		go func() {
			defer wg.Done()
			got, err := privateSpeedServerListWithNetwork(pst.NetworkIPv4)
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
	resetPrivateSpeedRegistryForTest(t, func(network pst.Network) (*pst.ServerList, error) {
		if network != pst.NetworkAuto {
			t.Errorf("registry network = %q, want automatic", network)
		}
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

func TestPrivateSpeedServerListSeparatesAddressFamilyCaches(t *testing.T) {
	var calls atomic.Int32
	registries := map[pst.Network]*pst.ServerList{
		pst.NetworkAuto: {TotalServers: 1},
		pst.NetworkIPv4: {TotalServers: 4},
		pst.NetworkIPv6: {TotalServers: 6},
	}
	resetPrivateSpeedRegistryForTest(t, func(network pst.Network) (*pst.ServerList, error) {
		calls.Add(1)
		return registries[network], nil
	})

	for _, network := range []pst.Network{pst.NetworkIPv4, pst.NetworkIPv6, pst.NetworkIPv4, pst.NetworkAuto, pst.NetworkIPv6} {
		got, err := privateSpeedServerListWithNetwork(network)
		if err != nil {
			t.Fatalf("load %q: %v", network, err)
		}
		if got != registries[network] {
			t.Fatalf("registry for %q = %#v, want %#v", network, got, registries[network])
		}
	}
	if got := calls.Load(); got != 3 {
		t.Fatalf("registry loader calls = %d, want one per address family", got)
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

func TestSelectPrivateSpeedCandidatesUsesDistinctCityFallbacks(t *testing.T) {
	candidates := []pst.ServerWithLatencyInfo{
		{Server: pst.ServerConfig{ID: "beijing-first", City: "Beijing"}},
		{Server: pst.ServerConfig{ID: "beijing-second", City: " beijing "}},
		{Server: pst.ServerConfig{ID: "shanghai", City: "Shanghai"}},
		{Server: pst.ServerConfig{ID: "guangzhou", City: "Guangzhou"}},
	}
	selected := selectPrivateSpeedCandidates(candidates, 2)
	if len(selected) != 4 {
		t.Fatalf("selected %d candidates, want all 4 fallbacks: %+v", len(selected), selected)
	}
	for index, want := range []string{"beijing-first", "shanghai", "guangzhou", "beijing-second"} {
		if selected[index].Server.ID != want {
			t.Fatalf("selected[%d] = %q, want %q", index, selected[index].Server.ID, want)
		}
	}
}

func TestSelectPrivateSpeedCandidatesCapsAttemptsAtTwiceLimit(t *testing.T) {
	candidates := make([]pst.ServerWithLatencyInfo, 0, 8)
	for index := 0; index < 8; index++ {
		candidates = append(candidates, pst.ServerWithLatencyInfo{Server: pst.ServerConfig{ID: fmt.Sprintf("node-%d", index), City: fmt.Sprintf("city-%d", index)}})
	}
	selected := selectPrivateSpeedCandidates(candidates, 2)
	if len(selected) != 4 {
		t.Fatalf("selected %d candidates, want 2*limit = 4", len(selected))
	}
}
