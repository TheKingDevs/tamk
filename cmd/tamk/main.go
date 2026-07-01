package main

import (
	"fmt"
	"os"

	"github.com/TheKingDevs/tamk/internal/delivery/cli"
	"github.com/TheKingDevs/tamk/pkg/logger"
)

func main() {
	cmd := cli.NewRootCmd()
	if err := cmd.Execute(); err != nil {
		logger.Error(fmt.Sprintf("Fatal error: %v", err))
		os.Exit(1)
	}
}
