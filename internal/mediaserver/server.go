package mediaserver

import (
	"context"

	"github.com/csnewman/cathode/internal/db"
	"golang.org/x/exp/slog"
)

type Server struct {
	logger *slog.Logger
	db     *db.DB
}

func New(logger *slog.Logger) (*Server, error) {
	db, err := db.Open("data/db.sqlite3")
	if err != nil {
		return nil, err
	}

	return &Server{
		logger: logger,
		db:     db,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	m, err := db.NewMigrator(ctx, s.logger, s.db)
	if err != nil {
		return err
	}

	version, err := m.CurrentVersion(ctx)
	if err != nil {
		return err
	}

	s.logger.Debug("Migration start", "current", version)

	return nil
}
