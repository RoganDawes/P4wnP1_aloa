package datastore

import (
	"errors"
	"github.com/dgraph-io/badger"
	_ "github.com/dgraph-io/badger"
)

/*
https://github.com/dgraph-io/badger --> Apache 2 (compatible)
 */

var (
	ErrCreate = errors.New("Error creating store")
	ErrOpen = errors.New("Error opening store")
	ErrGet = errors.New("Error retrieving value from store")
	ErrDelete = errors.New("Error deleting value from store")
	ErrPut = errors.New("Error putting value into store")
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
	s.Db,err = badger.Open(badgerOpts)
	if s.serializer == nil {
		s.serializer = NewSerializerProtobuf(false)
	}
	return err
}

func (s *Store) Close() {
	s.Db.Close()
}

func (s *Store) Put(key string, value interface{}, allowOverwrite bool) (err error) {
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

func (s *Store) Delete(key string) (err error) {
	err = s.Db.Update(func(txn *badger.Txn) error {
		// Your code here…
		return nil
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
