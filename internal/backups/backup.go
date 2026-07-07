package backups

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/MisterLaker/arangocado/internal/storages"
)

type Backup struct {
	Name        string
	CacheDir    string
	HistorySize int
	Storages    []string
	Database    Database
	Registry    storages.Registry
}

type DumpOptions struct {
	Storages     []string
	SkipSchedule bool
}

func (b *Backup) Dump(ctx context.Context, options *DumpOptions) error {
	t := time.Now()
	ts := t.Format("20060102T1504")

	cacheDir := b.GetCacheDir(ts)

	if err := b.Database.Dump(ctx, cacheDir); err != nil {
		return err
	}

	if err := b.Upload(ctx, t, ts, cacheDir, options); err != nil {
		return err
	}

	if err := b.RemoveCache(cacheDir); err != nil {
		return err
	}

	return nil
}

func (b *Backup) GetCacheDir(key string) string {
	// /tmp/<name>_<db>-<key>
	return fmt.Sprintf("%s/%s_%s-%s", b.CacheDir, b.Name, b.Database.GetName(), key)
}

type RestoreOptions struct {
	Key      string
	CacheDir string
	Database string
	Storage  string
}

func (b *Backup) Restore(ctx context.Context, options *RestoreOptions) error {
	if err := b.Download(ctx, options.Storage, options.Key); err != nil {
		return err
	}

	options.CacheDir = filepath.Join(b.CacheDir, storages.GetLocalPath(options.Key))

	if err := b.Database.Restore(ctx, options); err != nil {
		return err
	}

	if err := b.RemoveCache(options.CacheDir); err != nil {
		return err
	}

	return nil
}

func (b *Backup) RemoveCache(cacheDir string) error {
	return os.RemoveAll(cacheDir)
}

func listFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, info.Name())
		}
		return nil
	})

	return files, err
}

func (b *Backup) Upload(ctx context.Context, t time.Time, key, cacheDir string, options *DumpOptions) error {
	files, err := listFiles(cacheDir)
	if err != nil {
		return err
	}

	slog.Info("upload files", slog.String("key", key), slog.Int("files", len(files)))

	storages := options.Storages

	if len(storages) == 0 {
		storages = b.Storages
	}

	for _, s := range storages {
		storage, err := b.Registry.Get(s)
		if err != nil {
			return err
		}

		if !options.SkipSchedule && !storage.CanUpload(t) {
			slog.Info("skip backup upload", slog.String("backup", b.Name), slog.String("storage", s))

			continue
		}

		if err := storage.Upload(ctx, b.Name, b.Database.GetName(), key, cacheDir, files); err != nil {
			return err
		}
	}

	return nil
}

func (b *Backup) Download(ctx context.Context, storage, key string) error {
	s, err := b.Registry.Get(storage)
	if err != nil {
		return err
	}

	files, err := s.GetList(ctx, key)
	if err != nil {
		return err
	}

	slog.Info("downloading", slog.String("key", key), slog.Int("files", len(files)))

	return s.Download(ctx, b.CacheDir, files)
}

func (b *Backup) Remove(ctx context.Context, storageName string, keys []string) error {
	storage, err := b.Registry.Get(storageName)
	if err != nil {
		return err
	}

	for _, k := range keys {
		files, err := storage.GetList(ctx, k)
		if err != nil {
			return err
		}

		slog.Info("removing", slog.String("key", k), slog.Int("files", len(files)))

		if err := storage.Remove(ctx, files); err != nil {
			return err
		}
	}

	return nil
}

func pprint(items []string) {
	for _, k := range items {
		slog.Info("item", slog.String("name", k))
	}
}

func (b *Backup) getStorageDirPrefix() string {
	return fmt.Sprintf("%s/%s-", b.Name, b.Database.GetName())
}

func (b *Backup) CleanUp(ctx context.Context, storages ...string) error {
	key := b.getStorageDirPrefix()

	if len(storages) == 0 {
		storages = b.Storages
	}

	for _, s := range storages {
		storage, err := b.Registry.Get(s)
		if err != nil {
			return err
		}

		backups, err := storage.GetList(ctx, key)
		if err != nil {
			return err
		}

		// sort by date
		sort.Sort(sort.Reverse(sort.StringSlice(backups)))

		var removeList []string

		if len(backups) > b.HistorySize {
			removeList = backups[b.HistorySize:]
		}

		slog.Info("backups in storage", slog.String("storage", s))
		pprint(backups)

		if len(removeList) > 0 {
			slog.Info("backups to remove", slog.String("storage", s))
			pprint(removeList)

			if err := b.Remove(ctx, s, removeList); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Backup) List(ctx context.Context, storages ...string) error {
	if len(storages) == 0 {
		storages = b.Storages
	}

	key := b.getStorageDirPrefix()

	g, ctx := errgroup.WithContext(ctx)

	for _, storageName := range storages {
		storageName := storageName
		storage, err := b.Registry.Get(storageName)
		if err != nil {
			return err
		}

		g.Go(func() error {
			backups, err := storage.GetList(ctx, key)
			if err != nil {
				return err
			}

			// sort by date
			sort.Sort(sort.Reverse(sort.StringSlice(backups)))

			slog.Info("storage", slog.String("storage", storageName))
			pprint(backups)

			return nil
		})
	}

	return g.Wait()
}
