package main

import (
	"errors"

	"github.com/syndtr/goleveldb/leveldb"
)

func StoreGet(db *leveldb.DB, msg *StoreMessage) (string, error) {
	if db != nil {
		data, err := db.Get([]byte(msg.Key()), nil)
		if err != nil {
			return "", err
		}
		return string(data), nil
	} else {
		return "", errors.New("db is nil")
	}
}

func StoreSet(db *leveldb.DB, msg *StoreMessage) error {
	if db != nil {
		err := db.Put([]byte(msg.Key()), []byte(msg.Value()), nil)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("db is nil")
	}
}

func StoreHas(db *leveldb.DB, msg *StoreMessage) (bool, error) {
	if db != nil {
		ret, err := db.Has([]byte(msg.Key()), nil)
		if err != nil {
			return false, err
		}
		return ret, nil
	} else {
		return false, errors.New("db is nil")
	}
}

func StoreDelete(db *leveldb.DB, msg *StoreMessage) error {
	if db != nil {
		err := db.Delete([]byte(msg.Key()), nil)
		if err != nil {
			return err
		}
		return nil
	} else {
		return errors.New("db is nil")
	}
}
