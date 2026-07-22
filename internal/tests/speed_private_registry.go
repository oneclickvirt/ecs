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
	privateSpeedRegistryOnce   sync.Once
	privateSpeedRegistry       *pst.ServerList
	privateSpeedRegistryErr    error
	privateSpeedRegistryLoader = loadPrivateSpeedRegistry
)

func loadPrivateSpeedRegistry() (*pst.ServerList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	loaded, err := pst.LoadServerListWithMetadataContext(ctx)
	if err != nil || loaded.List == nil {
		return nil, fmt.Errorf("speedtest registry unavailable")
	}
	return loaded.List, nil
}

func privateSpeedServerList() (*pst.ServerList, error) {
	privateSpeedRegistryOnce.Do(func() {
		privateSpeedRegistry, privateSpeedRegistryErr = privateSpeedRegistryLoader()
	})
	return privateSpeedRegistry, privateSpeedRegistryErr
}
