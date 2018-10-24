package datastore

import (
	"errors"
	"github.com/dgraph-io/badger"
	_ "github.com/dgraph-io/badger"
	"os"
	"strings"
)

/*
https://github.com/dgraph-io/badger --> Apache 2 (compatible)
 */

var (
	ErrCreate = errors.New("Error creating store")
	ErrOpen = errors.New("Error opening store")
	ErrGet = errors.New("Error retrieving value from store")
	ErrKeys = errors.New("Error retrieving keys from store")
	ErrDelete = errors.New("Error deleting value from store")
	ErrPut = errors.New("Error putting value into store")
	ErrExists = errors.New("Error key exists already")
)


type Store struct {
	Path string
	Db *badger.DB
	serializer Serializer
}

func (s *Store) Open() (err error) {
	badgerOpts := badger.DefaultOptions
	badgerOpts.Dir = s.Path
	badgerOpts.ValueDir = s.Path
	badgerOpts.SyncWrites = true
	badgerOpts.ManagedTxns = true //needed for DropAll()
	s.Db,err = badger.Open(badgerOpts)
	if s.serializer == nil {
		s.serializer = NewSerializerProtobuf(false)
	}
	return err
}

func (s *Store) Close() {
	s.Db.Close()
}

// ToDo: Backup and restore could be synchronized to avoid concurrent transactions
func (s *Store) Backup(filePath string) (err error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer f.Close()
	_,err = s.Db.Backup(f,0)
	return
}

func (s *Store) Restore(filePath string) (err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	// DB should be cleared first
	err = s.Db.DropAll()
	if err != nil { return }
	err = s.Db.Load(f)
	return
}

func (s *Store) Put(key string, value interface{}, allowOverwrite bool) (err error) {
	// ToDo: Remove race condition (existence check and value set have to be merged into a single transaction)
	if !allowOverwrite && s.Exists(key) {
		return ErrExists
	}

	// serialize value
	sv,err := s.serializer.Encode(value)
	if err != nil { return }

	err = s.Db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(sv))
		return err
	})
	if err != nil { return ErrPut }
	return
}

func (s *Store) Exists(key string) (exists bool) {
	bkey := []byte(key)
	err := s.Db.View(func(txn *badger.Txn) error {
		_, err := txn.Get(bkey)
		if err != nil {
			return ErrGet
		}
		return nil
	})
	if err != nil { return false }
	return true
}

func (s *Store) Get(key string, target interface{}) (err error) {
	err = s.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}
		return s.serializer.Decode(val, target)
		//fmt.Printf("The answer is: %s\n", val)
		//return nil
	})
	if err != nil { return ErrGet }
	return
}

func (s *Store) Keys() (keys []string, err error) {
	err = s.Db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			keys = append(keys, string(iter.Item().Key()))
		}
		return nil
	})
	if err != nil { return keys, ErrKeys }
	return
}

func (s *Store) KeysPrefix(prefix string, trimPrefix bool) (keys []string, err error) {
	bprefix := []byte(prefix)
	err = s.Db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Seek(bprefix); iter.ValidForPrefix(bprefix); iter.Next() {
			if trimPrefix {
				s := strings.TrimPrefix(string(iter.Item().Key()), prefix)
				keys = append(keys,s)
			} else {
				keys = append(keys, string(iter.Item().Key()))
			}
		}
		return nil
	})
	if err != nil { return keys, ErrKeys }
	return
}

func (s *Store) Delete(key string) (err error) {
	err = s.Db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	if err != nil { return ErrDelete }
	return
}


func Open(path string) (store *Store, err error) {
	store = &Store{
		Path: path,
	}
	if err = store.Open(); err != nil {
		return nil,err
	}
	return
}
