package mediaserver

import (
	"context"
	"fmt"

	"github.com/csnewman/cathode/internal/db"
)

func (s *Server) migrate(ctx context.Context) error {
	m, err := db.NewMigrator(ctx, s.logger, s.db)
	if err != nil {
		return err
	}

	m.Register(1, s.migrateInit)

	return m.Migrate(ctx)
}

func (s *Server) migrateInit(_ context.Context, tx db.WTx) error {
	err := tx.Exec(`
		CREATE TABLE dsdm_servers (
			server     TEXT NOT NULL,
			CONSTRAINT dsdm_servers_pk PRIMARY KEY (server)
		) STRICT;
	`)
	if err != nil {
		return fmt.Errorf("failed to create dsdm_servers table: %w", err)
	}

	err = tx.Exec(`
		CREATE TABLE dsdm_certs (
			domain     TEXT NOT null,
			server     TEXT NOT NULL,
			cert       BLOB NOT NULL,
			pri_key    BLOB NOT NULL,
			issue_date TEXT NOT NULL,
			expire_date TEXT NOT NULL,
			CONSTRAINT dsdm_certs_pk PRIMARY KEY (domain)
		) STRICT;
	`)
	if err != nil {
		return fmt.Errorf("failed to create dsdm_certs table: %w", err)
	}

	return nil
}
