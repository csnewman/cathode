package main

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"github.com/csnewman/cathode/internal/mediaserver"
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

	if err := mainErr(logger); err != nil {
		logger.Error("Server Crashed", "err", err)
	} else {
		logger.Error("Server Stopped")
	}
}

func mainErr(logger *slog.Logger) error {
	logger.Info("Cathode Media Server")

	ms, err := mediaserver.New(logger)
	if err != nil {
		return err
	}

	ctx := context.Background()

	return ms.Run(ctx)
}
