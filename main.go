package main

import (
	. "github.com/oneclickvirt/ecs/defaultset"
)

func main() {
	InitLogger()
	defer Logger.Sync()
	Logger.Info("Start logging")
	// Logger.Info("Your log message", zap.Any("key", value))
}
