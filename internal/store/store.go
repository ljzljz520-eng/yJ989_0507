package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"activityregistration/internal/domain"
	bolt "go.etcd.io/bbolt"
)

var ErrNotFound = errors.New("record not found")

var bucketNames = [][]byte{
	[]byte("events"),
	[]byte("registrations"),
	[]byte("reviews"),
	[]byte("audits"),
	[]byte("batches"),
	[]byte("exports"),
}

type Store struct {
	db   *bolt.DB
	path string
}

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("store path is required")
	}
	if directory := filepath.Dir(path); directory != "." {
		if err := os.MkdirAll(directory, 0750); err != nil {
			return nil, err
		}
	}
	db, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	result := &Store{db: db, path: path}
	if err := result.ensureBuckets(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return result, nil
}

func (store *Store) ensureBuckets() error {
	return store.db.Update(func(transaction *bolt.Tx) error {
		for _, name := range bucketNames {
			if _, err := transaction.CreateBucketIfNotExists(name); err != nil {
				return err
			}
		}
		return nil
	})
}

func (store *Store) Close() error {
	if store == nil || store.db == nil {
		return nil
	}
	err := store.db.Close()
	store.db = nil
	return err
}

func (store *Store) Path() string {
	return store.path
}

func encode(value any) ([]byte, error) {
	return json.Marshal(value)
}

func decode(data []byte, target any) error {
	if len(data) == 0 {
		return ErrNotFound
	}
	return json.Unmarshal(data, target)
}

func (store *Store) put(bucket []byte, key string, value any) error {
	data, err := encode(value)
	if err != nil {
		return err
	}
	return store.db.Update(func(transaction *bolt.Tx) error {
		return transaction.Bucket(bucket).Put([]byte(key), data)
	})
}

func (store *Store) get(bucket []byte, key string, target any) error {
	return store.db.View(func(transaction *bolt.Tx) error {
		data := transaction.Bucket(bucket).Get([]byte(key))
		if data == nil {
			return ErrNotFound
		}
		return decode(data, target)
	})
}

func (store *Store) delete(bucket []byte, key string) error {
	return store.db.Update(func(transaction *bolt.Tx) error {
		return transaction.Bucket(bucket).Delete([]byte(key))
	})
}

func list[T any](store *Store, bucket []byte, target func([]byte) (T, error)) ([]T, error) {
	items := make([]T, 0)
	err := store.db.View(func(transaction *bolt.Tx) error {
		return transaction.Bucket(bucket).ForEach(func(key, value []byte) error {
			if value == nil {
				return nil
			}
			item, err := target(value)
			if err != nil {
				return err
			}
			items = append(items, item)
			return nil
		})
	})
	return items, err
}

func (store *Store) PutEvent(event domain.Event) error {
	return store.put(bucketNames[0], event.ID, event)
}

func (store *Store) GetEvent(id string) (domain.Event, error) {
	var event domain.Event
	err := store.get(bucketNames[0], id, &event)
	return event, err
}

func (store *Store) DeleteEvent(id string) error {
	return store.delete(bucketNames[0], id)
}

func (store *Store) ListEvents() ([]domain.Event, error) {
	return list(store, bucketNames[0], func(data []byte) (domain.Event, error) {
		var event domain.Event
		return event, decode(data, &event)
	})
}
