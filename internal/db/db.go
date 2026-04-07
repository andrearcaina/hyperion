package db

import (
	badger "github.com/dgraph-io/badger/v4"
)

type DB struct {
	db *badger.DB // embedded database
}

func New(path string) (*DB, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

func (d *DB) Get(key []byte) ([]byte, error) {
	var val []byte

	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		val, err = item.ValueCopy(val)
		return err
	})

	return val, err
}

func (d *DB) Set(key, val []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, val)
	})
}

func (d *DB) Delete(key []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (d *DB) NewTransaction(update bool) *badger.Txn {
	return d.db.NewTransaction(update)
}

func (d *DB) Close() error {
	return d.db.Close()
}
