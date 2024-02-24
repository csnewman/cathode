package db

import (
	"context"

	"golang.org/x/exp/slog"
)

type Migrator struct {
	logger *slog.Logger
	db     *DB
}

func NewMigrator(ctx context.Context, logger *slog.Logger, db *DB) (*Migrator, error) {
	err := db.Write(ctx, func(ctx context.Context, tx WTx) error {
		return tx.Exec(`
			CREATE TABLE IF NOT EXISTS migrations (
				id INTEGER PRIMARY KEY,
				applied TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
		`)
	})

	if err != nil {
		return nil, err
	}

	return &Migrator{
		logger,
		db,
	}, nil
}

func (m *Migrator) CurrentVersion(ctx context.Context) (int, error) {
	var id int

	err := m.db.Read(ctx, func(ctx context.Context, tx RTx) error {
		return tx.
			QueryRow(`SELECT MAX(id) FROM migrations`).
			Scan(&id)
	})

	return id, err
}
