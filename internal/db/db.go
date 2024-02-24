package db

import (
	"context"

	"github.com/tailscale/sqlite/sqliteh"
	"github.com/tailscale/sqlite/sqlitepool"
)

type Row = sqlitepool.Row

type Rows = sqlitepool.Rows

type DB struct {
	pool *sqlitepool.Pool
}

type RTx interface {
	QueryRow(sql string, args ...any) *Row

	Query(sql string, args ...any) (*Rows, error)
}

type WTx interface {
	RTx

	Exec(sql string, args ...any) error

	ExecRes(sql string, args ...any) (rowsAffected int64, err error)

	QueryRow(sql string, args ...any) *Row
}

func Open(path string) (*DB, error) {
	pool, err := sqlitepool.NewPool(
		path,
		4,
		func(d sqliteh.DB) error { return nil },
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &DB{
		pool,
	}, nil
}

func (d *DB) Close() error {
	return d.pool.Close()
}

func (d *DB) Write(ctx context.Context, fn func(ctx context.Context, tx WTx) error) error {
	tx, err := d.pool.BeginTx(ctx, "db::write")
	if err != nil {
		return err
	}

	defer tx.Rollback()

	if err := fn(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) Read(ctx context.Context, fn func(ctx context.Context, tx RTx) error) error {
	rx, err := d.pool.BeginRx(ctx, "db::read")
	if err != nil {
		return err
	}

	defer rx.Rollback()

	return fn(ctx, rx)
}
