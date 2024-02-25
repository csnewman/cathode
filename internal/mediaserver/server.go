package mediaserver

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/csnewman/cathode/internal/db"
)

type Server struct {
	logger  *slog.Logger
	db      *db.DB
	network *NetworkManager
}

func New(logger *slog.Logger) (*Server, error) {
	db, err := db.Open("data/db.sqlite3")
	if err != nil {
		return nil, err
	}

	nm := NewNetworkManager(logger, db)

	return &Server{
		logger:  logger,
		db:      db,
		network: nm,
	}, nil
}

func (s *Server) Run(ctx context.Context) error {
	if err := s.migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate db: %w", err)
	}

	s.network.Refresh(ctx)

	s.network.Run(ctx)

	return nil
}
