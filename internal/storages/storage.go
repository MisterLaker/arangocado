package storages

import (
	"context"
	"errors"
)

type Storage interface {
	Upload(ctx context.Context, backup, db, key, cacheDir string, files []string) error
	Download(ctx context.Context, cacheDir string, files []string) error
	GetList(ctx context.Context, prefix string) ([]string, error)
	Remove(ctx context.Context, files []string) error
}

type CombinedStorage struct {
	Storages []Storage
}

func (cs *CombinedStorage) Upload(ctx context.Context, backup, db, key, cacheDir string, files []string) error {
	for _, s := range cs.Storages {
		if err := s.Upload(ctx, backup, db, key, cacheDir, files); err != nil {
			return err
		}
	}

	return nil
}

func (cs *CombinedStorage) Download(ctx context.Context, cacheDir string, files []string) error {
	for _, s := range cs.Storages {
		if err := s.Download(ctx, cacheDir, files); err != nil {
			return err
		}
	}

	return nil
}

func (cs *CombinedStorage) GetList(ctx context.Context, prefix string) ([]string, error) {
	return cs.Storages[0].GetList(ctx, prefix)
}

func (cs *CombinedStorage) Remove(ctx context.Context, files []string) error {
	for _, s := range cs.Storages {
		if err := s.Remove(ctx, files); err != nil {
			return err
		}
	}

	return nil
}

type Registry map[string]Storage

var (
	ErrNotFound = errors.New("storage not found")
	ErrNotDefined = errors.New("storages are not defined")
)

func (r Registry) Get(keys []string) (Storage, error) {
	var items []Storage

	for _, k := range keys {
		s, ok := r[k]
		if !ok {
			return nil, ErrNotFound
		}

		items = append(items, s)
	}

	if len(items) == 0 {
		return nil, ErrNotDefined
	}

	return &CombinedStorage{Storages: items}, nil
}
