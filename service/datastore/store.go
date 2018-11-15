// +build linux,arm

package datastore

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/dgraph-io/badger/options"
	"io/ioutil"
	"os"
	"strings"
)

/*
https://github.com/dgraph-io/badger --> Apache 2 (compatible)
 */

var (
	ErrCreate = errors.New("Error creating store")
	ErrOpen   = errors.New("Error opening store")
	ErrGet    = errors.New("Error retrieving value from store")
	ErrKeys   = errors.New("Error retrieving keys from store")
	ErrDelete = errors.New("Error deleting value from store")
	ErrPut    = errors.New("Error putting value into store")
	ErrExists = errors.New("Error key exists already")
)

type Store struct {
	Path       string
	Db         *badger.DB
	serializer Serializer
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func (s *Store) Open(initDbBackupPath string) (err error) {
	badgerOpts := badger.DefaultOptions
	badgerOpts.Dir = s.Path
	badgerOpts.ValueDir = s.Path
	badgerOpts.SyncWrites = true
	badgerOpts.TableLoadingMode = options.FileIO
	badgerOpts.ValueLogLoadingMode = options.FileIO

	// check if DB dir exists
	exists,err := exists(s.Path)
	if err != nil { return err }

	s.Db, err = badger.Open(badgerOpts)
	if s.serializer == nil {
		s.serializer = NewSerializerProtobuf(false)
	}

	//If the s.Path didn't exist, we have a clean and empty DB at this point and thus restore a initial db
	if !exists {
		err = s.Restore(initDbBackupPath, true)
	}

	return err
}

func (s *Store) Close() {
	s.Db.Close()
}

func (s *Store) Clear() (err error) {
	keys, err := s.Keys()
	if err != nil {
		return
	}
	// ToDo: Other transactions could add/delete keys meanwhile, which should be avoided (without locking the functions accessing badger)
	err = s.DeleteMulti(keys)

	//s.Db.DropAll()

	return nil
}

// ToDo: Backup and restore could be synchronized to avoid concurrent transactions
func (s *Store) Backup(filePath string) (err error) {
	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0664)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = s.Db.Backup(f, 0)
	return
}

func (s *Store) Restore(filePath string, replace bool) (err error) {
	fmt.Println("Restoring DB...")

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if replace {
		// As backups are version agnostic, we can't replace newer keys or keys deleted in a later version
		// To solve this, replacement is handled like this
		// - the backup is loaded into a temporary DB which hasn't got any (versioned) entries
		// - all keys of the current DB are deleted (deletion has results in a newer version of the key as the one in the backup)
		// - copy over all key-value-pairs of the temporary DB to the current DB via set (updates the version of the key)
		fmt.Println("...Clear current DB")
		err = s.Clear()
		if err != nil {
			return
		}

		//create temp db
		fmt.Println("... create temp DB")
		tmpDbDir, err := ioutil.TempDir("/tmp", "badger_backup")
		if err != nil {
			return err
		}
		fmt.Println("... temp DB created")
		defer os.RemoveAll(tmpDbDir)

		fmt.Println("... opening temp DB at", tmpDbDir)
		badgerOpts := badger.DefaultOptions
		badgerOpts.Dir = tmpDbDir
		badgerOpts.ValueDir = tmpDbDir
		badgerOpts.SyncWrites = true
		badgerOpts.TableLoadingMode = options.FileIO
		badgerOpts.ValueLogLoadingMode = options.FileIO
		tmpDB, err := badger.Open(badgerOpts)
		if err != nil {
			fmt.Println("... opening of temp DB failed:", err)
			return err
		}
		defer tmpDB.Close()

		fmt.Println("... loading backup to temp DB")
		tmpDB.Load(f)

		fmt.Println("... retrieving keys of temp DB")
		restoreKeys := make([][]byte, 0)
		err = tmpDB.View(func(txn *badger.Txn) error {
			iterOpts := badger.DefaultIteratorOptions
			iterOpts.PrefetchValues = false
			iter := txn.NewIterator(iterOpts)
			defer iter.Close()
			for iter.Rewind(); iter.Valid(); iter.Next() {
				key := iter.Item().Key()
				fmt.Println("... found key:", string(key))
				resKey := make([]byte, len(key))
				copy(resKey, key)
				restoreKeys = append(restoreKeys, resKey)
			}
			return nil
		})
		if err != nil {
			return err
		}

		keycount := len(restoreKeys)
		fmt.Printf("... found %d keys", keycount)

		txnReadSrc := tmpDB.NewTransaction(false)
		defer txnReadSrc.Discard()

		fmt.Printf("... restoring %d key-value-pairs\n", keycount)
		for idx, restoreKey := range restoreKeys {
			fmt.Printf("... copy over key %d of %d '%s' ...", idx+1, keycount, string(restoreKey))
			item, err := txnReadSrc.Get(restoreKey)
			if err != nil {
				continue
			}
			s.Db.Update(func(txn *badger.Txn) error {
				return item.Value(func(val []byte) error {

					// Ignore keys with empty vals, see https://github.com/dgraph-io/badger/issues/521
					if len(val) > 0 {
						txn.Set(restoreKey, val)
						fmt.Println("added")
					} else {
						fmt.Println("ignored, empty value")
					}
					return nil
				})
			})
		}

		return nil

	} else {
		err = s.Db.Load(f)
		return
	}

}

func (s *Store) Put(key string, value interface{}, allowOverwrite bool) (err error) {
	// ToDo: Remove race condition (existence check and value set have to be merged into a single transaction)
	if !allowOverwrite && s.Exists(key) {
		return ErrExists
	}

	// serialize value
	sv, err := s.serializer.Encode(value)
	if err != nil {
		return
	}

	err = s.Db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(key), []byte(sv))
		return err
	})
	if err != nil {
		return ErrPut
	}
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
	if err != nil {
		return false
	}
	return true
}

func (s *Store) Get(key string, target interface{}) (err error) {
	err = s.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		val,err := item.ValueCopy([]byte{})
		if err != nil {
			return err
		}
		return s.serializer.Decode(val, target)
		//fmt.Printf("The answer is: %s\n", val)
		//return nil
	})
	if err != nil {
		return ErrGet
	}
	return
}

func (s *Store) Keys() (keys []string, err error) {
	err = s.Db.View(func(txn *badger.Txn) error {
		iterOpts := badger.DefaultIteratorOptions
		iterOpts.PrefetchValues = false
		iter := txn.NewIterator(iterOpts)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			key := iter.Item().Key()
			resKey := make([]byte, len(key))
			copy(resKey, key)
			keys = append(keys, string(resKey))

		}
		return nil
	})
	if err != nil {
		return keys, ErrKeys
	}
	return
}

func (s *Store) KeysPrefix(prefix string, trimPrefix bool) (keys []string, err error) {
	bprefix := []byte(prefix)
	err = s.Db.View(func(txn *badger.Txn) error {
		iterOpts := badger.DefaultIteratorOptions
		iterOpts.PrefetchValues = false
		iter := txn.NewIterator(iterOpts)
		defer iter.Close()
		for iter.Seek(bprefix); iter.ValidForPrefix(bprefix); iter.Next() {
			key := iter.Item().Key()
			resKey := make([]byte, len(key))
			copy(resKey, key)
			if trimPrefix {
				s := strings.TrimPrefix(string(resKey), prefix)
				keys = append(keys, s)
			} else {
				keys = append(keys, string(resKey))
			}
		}
		return nil
	})
	if err != nil {
		return keys, ErrKeys
	}
	return
}

func (s *Store) Delete(key string) (err error) {
	err = s.Db.Update(func(txn *badger.Txn) error {

		return txn.Delete([]byte(key))
	})
	if err != nil {
		return ErrDelete
	}
	return
}

func (s *Store) DeleteMulti(keys []string) (err error) {
	txn := s.Db.NewTransaction(true)

	for _, key := range keys {
		errDel := txn.Delete([]byte(key))
		if errDel != nil {
			if err == badger.ErrTxnTooBig {
				txn.Commit()                 // commit current transaction
				txn = s.Db.NewTransaction(true) // replace with new transaction
				txn.Delete([]byte(key))         // add Delete which produced error to new transaction
			} else {
				txn.Discard()
				return ErrDelete
			}
		}
	}
	txn.Commit()
	txn.Discard()
	return nil
}

func Open(workPath string, initDbBackupPath string) (store *Store, err error) {
	store = &Store{
		Path: workPath,
	}
	if err = store.Open(initDbBackupPath); err != nil {
		return nil, err
	}
	return
}
