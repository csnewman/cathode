package main

import (
	"os"

	"github.com/csnewman/cathode/mediaserver"
	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))

	mediaserver.NewServer(logger)
}
