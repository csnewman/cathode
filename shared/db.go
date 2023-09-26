package shared

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"golang.org/x/exp/slog"
)

const DebugDB = false

type DB struct {
	logger *slog.Logger
	db     *badger.DB
}

func NewDB(logger *slog.Logger, path string) (*DB, error) {
	db, err := badger.Open(
		badger.DefaultOptions(path).
			WithLogger(&badgerLogger{
				logger: logger,
			}),
	)
	if err != nil {
		return nil, err
	}

	return &DB{
		logger: logger,
		db:     db,
	}, nil
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Transact(update bool, fn func(tx *Tx) error) error {
	txn := d.db.NewTransaction(update)
	defer txn.Discard()

	if err := fn(&Tx{logger: d.logger, txn: txn}); err != nil {
		return err
	}

	return txn.Commit()
}

type Tx struct {
	logger *slog.Logger
	txn    *badger.Txn
}

func (t *Tx) Get(key string, tgt any) error {
	val, err := t.GetRaw(key)
	if err != nil {
		return err
	}

	if val == nil {
		return nil
	}

	return json.Unmarshal(val, tgt)
}

func (t *Tx) Set(key string, value any) error {
	enc, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return t.SetRaw(key, enc)
}

func (t *Tx) GetRaw(key string) ([]byte, error) {
	item, err := t.txn.Get([]byte(key))
	if errors.Is(err, badger.ErrKeyNotFound) {
		if DebugDB {
			t.logger.Debug("GetRaw", "key", key)
		}

		return nil, nil
	} else if err != nil {
		return nil, err
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, err
	}

	if DebugDB {
		t.logger.Debug("GetRaw", "key", key, "value", value)
	}

	return value, nil
}

func (t *Tx) SetRaw(key string, value []byte) error {
	if DebugDB {
		t.logger.Debug("SetRaw", "key", key, "value", value)
	}

	return t.txn.Set([]byte(key), value)
}

func (t *Tx) Delete(key string) error {
	if DebugDB {
		t.logger.Debug("Delete", "key", key)
	}

	return t.txn.Delete([]byte(key))
}

type badgerLogger struct {
	logger *slog.Logger
}

func (b *badgerLogger) Errorf(s string, i ...interface{}) {
	for _, line := range strings.Split(strings.Trim(fmt.Sprintf(s, i...), "\n"), "\n") {
		b.logger.Error("Badger Error", "inner", line)
	}
}

func (b *badgerLogger) Warningf(s string, i ...interface{}) {
	for _, line := range strings.Split(strings.Trim(fmt.Sprintf(s, i...), "\n"), "\n") {
		b.logger.Warn("Badger Warning", "inner", line)
	}
}

func (b *badgerLogger) Infof(_ string, _ ...interface{}) {
}

func (b *badgerLogger) Debugf(_ string, _ ...interface{}) {
}
