package main

import (
	"os"

	"github.com/TheKingDevs/tamk/internal/delivery/cli"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		logger.Error("Fatal error", "error", err)
		os.Exit(1)
	}
}
