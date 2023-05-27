package main

import (
	"os"
	"strings"

	"github.com/csnewman/cathode/mediaserver"
	"golang.org/x/exp/slog"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) == 0 && a.Key == slog.SourceKey {
				//nolint:forcetypeassert
				source := a.Value.Any().(*slog.Source)
				source.File = strings.TrimPrefix(source.File, "github.com/csnewman/cathode/")
			}

			return a
		},
	}))

	logger.Info("Cathode MediaServer")

	ms := mediaserver.NewServer(logger)
	defer ms.Close()

	if err := ms.Run(); err != nil {
		logger.Error("Server Crashed", "err", err)
	} else {
		logger.Error("Server Stopped")
	}
}
