// Command data-sync refreshes the checked-in goecs data snapshot. It is kept
// separate from the runtime so a scheduled job can update only generated
// JSON and the manifest.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	datasync "github.com/oneclickvirt/ecs/internal/data/sync"
)

func main() {
	output := flag.String("output", datasync.DefaultOutputDir, "directory for generated data")
	showVersion := flag.Bool("version", false, "show data synchronizer version")
	timeout := flag.Duration("timeout", 5*time.Minute, "overall synchronization timeout")
	flag.Parse()
	if *showVersion {
		fmt.Println(datasync.Version())
		return
	}
	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "timeout must be positive")
		os.Exit(2)
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	changed, err := datasync.Sync(ctx, *output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sync data: %v\n", err)
		os.Exit(1)
	}
	if changed {
		fmt.Println("data snapshot updated")
	} else {
		fmt.Println("data snapshot unchanged")
	}
}
