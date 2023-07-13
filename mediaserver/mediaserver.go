package mediaserver

import (
	"context"
	"log"

	"github.com/csnewman/cathode/shared"
	"golang.org/x/exp/slog"
)

type Server struct {
	logger *slog.Logger
	db     *shared.DB
}

func NewServer(logger *slog.Logger) *Server {
	db, err := shared.NewDB(logger, "db")
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		logger: logger,
		db:     db,
	}
}

func (s *Server) Close() error {
	return s.db.Close()
}

func (s *Server) Run(ctx context.Context) error {
	nm, err := NewNetworkManager(s.logger, s.db)
	if err != nil {
		return err
	}

	nm.Refresh(ctx)

	nm.Run(ctx)

	return nil
}
