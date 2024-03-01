package backup

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MisterLaker/arangocado/internal/storages"
)

type Backup struct {
	Name        string
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	Collections []string
	CacheDir    string
	HistorySize int
	Storages    []string
	Registry    storages.Registry
}

func (b *Backup) Dump(ctx context.Context) error {
	ts := time.Now().Format("20060102T1504")
	cacheDir := b.GetCacheDir(ts)

	if err := b.ArangoDump(ctx, cacheDir); err != nil {
		return err
	}

	if err := b.Upload(ctx, ts, cacheDir); err != nil {
		return err
	}

	if err := b.RemoveCache(cacheDir); err != nil {
		return err
	}

	return nil
}

func (b *Backup) GetCacheDir(key string) string {
	// /tmp/<name>_<db>-<key>
	return fmt.Sprintf("%s/%s_%s-%s", b.CacheDir, b.Name, b.Database, key)
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

	options.CacheDir = filepath.Join(b.CacheDir, getLocalPath(options.Key))

	if err := b.ArangoRestore(ctx, options); err != nil {
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

func (b *Backup) ArangoDump(ctx context.Context, cacheDir string) error {
	args := map[string]any{
		"server.endpoint":            b.Host,
		"server.username":            b.User,
		"server.password":            b.Password,
		"server.database":            b.Database,
		"collection":                 b.Collections,
		"output-directory":           cacheDir,
		"overwrite":                  true,
		"include-system-collections": true,
	}

	cmd := exec.CommandContext(ctx, "arangodump", makeCommandArgs(args)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (b *Backup) ArangoRestore(ctx context.Context, options *RestoreOptions) error {
	db := b.Database
	if options.Database != "" {
		db = options.Database
	}

	args := map[string]any{
		"server.endpoint":            b.Host,
		"server.username":            b.User,
		"server.password":            b.Password,
		"server.database":            db,
		"collection":                 b.Collections,
		"input-directory":            options.CacheDir,
		"create-database":            true,
		"overwrite":                  true,
		"include-system-collections": true,
	}

	log.Println("arangorestore", "database:", db, "cache dir:", options.CacheDir)

	cmd := exec.CommandContext(ctx, "arangorestore", makeCommandArgs(args)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func makeCommandArgs(args map[string]any) []string {
	var items []string

	for k, v := range args {
		switch t := v.(type) {
		case bool:
			items = append(items, fmt.Sprintf("--%s", k))
		case []string:
			for _, kv := range t {
				items = append(items, fmt.Sprintf("--%s=%s", k, kv))
			}
		default:
			items = append(items, fmt.Sprintf("--%s=%v", k, v))
		}
	}

	return items
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

func (b *Backup) Upload(ctx context.Context, key, cacheDir string) error {
	files, err := listFiles(cacheDir)
	if err != nil {
		return err
	}

	log.Println("upload files ", key, "files: ", len(files))

	for _, s := range b.Storages {
		storage, err := b.Registry.Get(s)
		if err != nil {
			return err
		}

		if err := storage.Upload(ctx, b.Name, b.Database, key, cacheDir, files); err != nil {
			return err
		}
	}

	return nil
}

func getLocalPath(s string) string {
	return strings.Replace(s, "/", "_", 1)
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

	log.Println("downloading: ", key, "files: ", len(files))

	return s.Download(ctx, b.CacheDir, files)
}

func (b *Backup) Remove(ctx context.Context, storage string, keys []string) error {
	s, err := b.Registry.Get(storage)
	if err != nil {
		return err
	}

	for _, k := range keys {
		files, err := s.GetList(ctx, k)
		if err != nil {
			return err
		}

		log.Println("removing: ", k, "files: ", len(files))

		if err := s.Remove(ctx, files); err != nil {
			return err
		}
	}

	return nil
}

func pprint(items []string) {
	for _, k := range items {
		log.Printf("\t- %s\n", k)
	}
}

func (b *Backup) CleanUp(ctx context.Context) error {
	key := fmt.Sprintf("%s/%s-", b.Name, b.Database)

	for _, s := range b.Storages {
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

		log.Println("backups storage: ", s)
		pprint(backups)

		log.Println("backups to remove: ")
		pprint(removeList)

		if len(removeList) > 0 {
			if err := storage.Remove(ctx, removeList); err != nil {
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

	key := fmt.Sprintf("%s/%s-", b.Name, b.Database)

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

		log.Println("backups storage: ", s)
		pprint(backups)
	}

	return nil
}
