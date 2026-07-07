package storages

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"golang.org/x/sync/errgroup"
)

var ErrNotFound = errors.New("storage not found")

type Storage interface {
	Upload(ctx context.Context, backup, db, key, cacheDir string, files []string) error
	Download(ctx context.Context, cacheDir string, files []string) error
	GetList(ctx context.Context, prefix string) ([]string, error)
	Remove(ctx context.Context, files []string) error
	CanUpload(t time.Time) bool
}

type Registry map[string]Storage

func (r Registry) Get(key string) (Storage, error) {
	s, ok := r[key]
	if !ok {
		return nil, ErrNotFound
	}

	return s, nil
}

type ChunkStorage struct {
	Storage
	Workers int
}

func (s *ChunkStorage) Upload(ctx context.Context, backup, db, ts, cacheDir string, files []string) error {
	ln := len(files)
	chunkSize := (ln + s.Workers - 1) / s.Workers

	slog.Info("upload files", slog.Int("files", ln), slog.Int("chunk_size", chunkSize))

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < ln; i += chunkSize {
		i := i
		end := i + chunkSize

		if end > ln {
			end = ln
		}

		g.Go(func() error {
			return s.Storage.Upload(ctx, backup, db, ts, cacheDir, files[i:end])
		})
	}

	return g.Wait()
}

func (s *ChunkStorage) Download(ctx context.Context, cacheDir string, files []string) error {
	ln := len(files)
	chunkSize := (ln + s.Workers - 1) / s.Workers

	slog.Info("downloading files", slog.Int("files", ln), slog.Int("chunk_size", chunkSize))

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < ln; i += chunkSize {
		i := i
		end := i + chunkSize

		if end > ln {
			end = ln
		}

		g.Go(func() error {
			return s.Storage.Download(ctx, cacheDir, files[i:end])
		})
	}

	return g.Wait()
}

type ScheduleStorage struct {
	Storage
	Schedule     cron.Schedule
	nextUpdateAt time.Time
}

func NewScheduleStorage(storage Storage, schedule cron.Schedule, t time.Time) *ScheduleStorage {
	s := &ScheduleStorage{
		Storage:  storage,
		Schedule: schedule,
	}

	s.setNextUpdateAt(t)

	return s
}

func (s *ScheduleStorage) setNextUpdateAt(t time.Time) {
	s.nextUpdateAt = s.Schedule.Next(t)
}

func (s *ScheduleStorage) CanUpload(t time.Time) bool {
	return s.nextUpdateAt.Before(t)
}

func (s *ScheduleStorage) Upload(ctx context.Context, backup, db, ts, cacheDir string, files []string) error {
	t := time.Now()

	if err := s.Storage.Upload(ctx, backup, db, ts, cacheDir, files); err != nil {
		return err
	}

	s.setNextUpdateAt(t)

	return nil
}
