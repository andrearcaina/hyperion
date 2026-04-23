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

	return &DB{
		db: db,
	}, nil
}

func (d *DB) Get(key []byte) ([]byte, error) {
	var val []byte

	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		val, err = item.ValueCopy(nil)
		return err
	})

	return val, err
}

func (d *DB) Set(key, value []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func (d *DB) Delete(key []byte) error {
	return d.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(key)
	})
}

func (d *DB) ForEach(fn func(key, value []byte) error) error {
	txn := d.db.NewTransaction(false) // read only transaction
	defer txn.Discard()

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()

	// this is like doing for i := 0; i < len(db); i++ { ... }
	// but obviously we don't have a length of db, so we just iterate until there are no more items
	// the iterator will automatically move to the next item when we call it.Next() (using badger's Iterator API)
	for it.Rewind(); it.Valid(); it.Next() {
		item := it.Item()

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		// call the provided function with the key and value
		// could be printing them, or processing, etc.
		if err := fn(item.Key(), val); err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) Clear() error {
	return d.db.DropAll()
}

func (d *DB) Close() error {
	return d.db.Close()
}
