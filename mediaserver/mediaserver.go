package mediaserver

import (
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

type Test struct {
	Server string `json:"server"`
	Made   bool   `json:"made"`
}

func (s *Server) Run() error {
	return s.db.Transact(true, func(tx *shared.Tx) error {
		var test Test

		if err := tx.Get("test", &test); err != nil {
			s.logger.Error("test", "e", err)

			return err
		}

		test = Test{
			Server: "123",
			Made:   true,
		}

		if err := tx.Set("test", test); err != nil {
			s.logger.Error("test-set", "e", err)

			return err
		}

		return nil
	})
}
