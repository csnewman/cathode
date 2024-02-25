package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var ErrDBTooNew = errors.New("database version newer than supported")

type MigrateFunc func(ctx context.Context, tx WTx) error

type Migrator struct {
	logger *slog.Logger
	db     *DB
	funcs  map[int]MigrateFunc
	maxVer int
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
		logger: logger,
		db:     db,
		funcs:  make(map[int]MigrateFunc),
	}, nil
}

func (m *Migrator) Register(version int, f MigrateFunc) {
	if _, ok := m.funcs[version]; ok {
		panic("version already in use")
	}

	m.funcs[version] = f

	if version > m.maxVer {
		m.maxVer = version
	}
}

func (m *Migrator) CurrentVersion(ctx context.Context) (int, error) {
	var id int

	err := m.db.Read(ctx, func(_ context.Context, tx RTx) error {
		return tx.
			QueryRow(`SELECT MAX(id) FROM migrations`).
			Scan(&id)
	})

	return id, err
}

func (m *Migrator) Migrate(ctx context.Context) error {
	curVer, err := m.CurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if curVer > m.maxVer {
		m.logger.Error("Database newer than supported", "file", curVer, "supported", m.maxVer)

		return ErrDBTooNew
	}

	if curVer == m.maxVer {
		m.logger.Info("Database is up-to-date")

		return nil
	}

	m.logger.Info("Database is outdated", "current", curVer, "target", m.maxVer)

	for i := curVer + 1; i <= m.maxVer; i++ {
		m.logger.Info("Applying migration", "version", i)

		err := m.db.Write(ctx, func(ctx context.Context, tx WTx) error {
			if err := m.funcs[i](ctx, tx); err != nil {
				return fmt.Errorf("apply func failed: %w", err)
			}

			err := tx.Exec(
				`INSERT INTO migrations (id, applied) VALUES ($1, datetime())`,
				i,
			)
			if err != nil {
				return fmt.Errorf("log insert failed: %w", err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to apply migration %v: %w", i, err)
		}
	}

	m.logger.Info("Database updated")

	return nil
}
