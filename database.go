package pirategopher

import (
	"errors"
	"github.com/boltdb/bolt"
	"log"
)

type BoltDB struct {
	*bolt.DB
}

var (
	errorBucketDoesNotExist = errors.New("bucket does not exist")
)

func openDb(name string) *BoltDB {
	db, err := bolt.Open(name, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	return &BoltDB{db}
}

func (db BoltDB) find(key, bucket string) (value string, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errorBucketDoesNotExist
		}

		value = string(b.Get([]byte(key)))
		return nil
	})
	return value, err
}

func (db BoltDB) delete(key, bucket string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errorBucketDoesNotExist
		}
		return b.Delete([]byte(key))
	})
	return err
}

// Delete a bucket
func (db BoltDB) deleteBucket(bucket string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucket))
	})
	return err
}

// Create or update a value
func (db *BoltDB) createOrUpdate(key, value, bucket string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		err = b.Put([]byte(key), []byte(value))
		return err
	})
}

// Check if a key is available
func (db BoltDB) isAvailable(key, bucket string) (bool, error) {
	available := false
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errorBucketDoesNotExist
		}

		v := b.Get([]byte(key))
		if v == nil {
			available = true
		}
		return nil
	})

	return available, err
}
