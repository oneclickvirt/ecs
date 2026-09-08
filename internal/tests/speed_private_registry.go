//go:build !ecs_public

package tests

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/oneclickvirt/privatespeedtest/pst"
)

var (
	privateSpeedRegistryMu     sync.Mutex
	privateSpeedRegistries     = make(map[pst.Network]*privateSpeedRegistryEntry)
	privateSpeedRegistryLoader = loadPrivateSpeedRegistry
)

type privateSpeedRegistryEntry struct {
	once sync.Once
	list *pst.ServerList
	err  error
}

func loadPrivateSpeedRegistry(network pst.Network) (*pst.ServerList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	loaded, err := pst.LoadServerListWithMetadataContextWithNetwork(ctx, network)
	if err != nil || loaded.List == nil {
		return nil, fmt.Errorf("speedtest registry unavailable")
	}
	return loaded.List, nil
}

func normalizedPrivateSpeedRegistryNetwork(network pst.Network) pst.Network {
	switch network {
	case pst.NetworkIPv4, pst.NetworkIPv6:
		return network
	default:
		return pst.NetworkAuto
	}
}

func privateSpeedServerListWithNetwork(network pst.Network) (*pst.ServerList, error) {
	network = normalizedPrivateSpeedRegistryNetwork(network)
	privateSpeedRegistryMu.Lock()
	entry := privateSpeedRegistries[network]
	if entry == nil {
		entry = &privateSpeedRegistryEntry{}
		privateSpeedRegistries[network] = entry
	}
	privateSpeedRegistryMu.Unlock()

	entry.once.Do(func() {
		entry.list, entry.err = privateSpeedRegistryLoader(network)
	})
	return entry.list, entry.err
}

// privateSpeedServerList preserves the historical automatic entrypoint for
// callers that have not selected a measurement address family.
func privateSpeedServerList() (*pst.ServerList, error) {
	return privateSpeedServerListWithNetwork(pst.NetworkAuto)
}
