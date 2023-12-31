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

	"github.com/minio/minio-go/v7"
	"golang.org/x/sync/errgroup"
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
	Bucket      string
	Workers     int
	HistorySize int

	Minio *minio.Client
}

func (b *Backup) Run(ctx context.Context) error {
	ts := time.Now().Format("20060102T1504")
	cacheDir := b.GetCacheDir(ts)

	if err := b.Arangodump(ctx, cacheDir); err != nil {
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
}

func (b *Backup) Restore(ctx context.Context, options *RestoreOptions) error {
	if err := b.Download(ctx, options.Key); err != nil {
		return err
	}

	options.CacheDir = filepath.Join(b.CacheDir, getLocalPath(options.Key))

	if err := b.Arangorestore(ctx, options); err != nil {
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

func (b *Backup) Arangodump(ctx context.Context, cacheDir string) error {
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

	cmd := exec.CommandContext(ctx, "arangodump", makeCmdArgs(args)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (b *Backup) Arangorestore(ctx context.Context, options *RestoreOptions) error {
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

	cmd := exec.CommandContext(ctx, "arangorestore", makeCmdArgs(args)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func makeCmdArgs(args map[string]any) []string {
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

func (b *Backup) UploadFiles(ctx context.Context, key, cacheDir string, files []string) error {
	for _, name := range files {
		objectName := fmt.Sprintf("%s/%s-%s/%s", b.Name, b.Database, key, name)
		path := filepath.Join(cacheDir, name)

		log.Println("file: ", path)
		log.Println("path: ", objectName)

		_, err := b.Minio.FPutObject(ctx, b.Bucket, objectName, path, minio.PutObjectOptions{})
		if err != nil {
			log.Println("upload error: ", err)

			continue
		}

		log.Println("Successfully uploaded file: ", path)
	}

	return nil
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

	ln := len(files)
	chunkSize := (ln + b.Workers - 1) / b.Workers

	log.Println("upload files ", "files: ", ln, "chunk_size: ", chunkSize)

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < ln; i += chunkSize {
		i := i
		end := i + chunkSize

		if end > ln {
			end = ln
		}

		g.Go(func() error {
			return b.UploadFiles(ctx, key, cacheDir, files[i:end])
		})
	}

	return g.Wait()
}

func getLocalPath(s string) string {
	return strings.Replace(s, "/", "_", 1)
}

func (b *Backup) DownloadFiles(ctx context.Context, files []string) error {
	for _, name := range files {
		localPath := getLocalPath(name)

		path := fmt.Sprintf("%s/%s", b.CacheDir, localPath)

		log.Println("file: ", path)
		log.Println("path: ", name)

		err := b.Minio.FGetObject(ctx, b.Bucket, name, path, minio.GetObjectOptions{})
		if err != nil {
			log.Println("download error: ", err)

			continue
		}

		log.Println("Successfully downloaded file: ", path)
	}

	return nil
}

func (b *Backup) Download(ctx context.Context, key string) error {
	files, err := b.listObjects(ctx, key)
	if err != nil {
		return err
	}

	ln := len(files)
	chunkSize := (ln + b.Workers - 1) / b.Workers

	log.Println("downloading: ", key, "files: ", ln, "chunk_size: ", chunkSize)

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < ln; i += chunkSize {
		i := i
		end := i + chunkSize

		if end > ln {
			end = ln
		}

		g.Go(func() error {
			return b.DownloadFiles(ctx, files[i:end])
		})
	}

	return g.Wait()
}

func (b *Backup) Remove(ctx context.Context, keys []string) error {
	for _, k := range keys {
		objects, err := b.listObjects(ctx, k)
		if err != nil {
			return err
		}

		log.Println("removing: ", k, "files: ", len(objects))

		for _, name := range objects {
			if err := b.Minio.RemoveObject(ctx, b.Bucket, name, minio.RemoveObjectOptions{}); err != nil {
				return err
			}
		}
	}

	return nil
}

func (b *Backup) listObjects(ctx context.Context, prefix string) ([]string, error) {
	objectCh := b.Minio.ListObjects(ctx, b.Bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: false,
	})

	var items []string

	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}

		items = append(items, object.Key)
	}

	return items, nil
}

func pprint(items []string) {
	for _, k := range items {
		log.Printf("\t- %s\n", k)
	}
}

func (b *Backup) CleanUp(ctx context.Context) error {
	backups, err := b.listObjects(ctx, fmt.Sprintf("%s/%s-", b.Name, b.Database))
	if err != nil {
		return err
	}

	// sort by date
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	var removeList []string

	if len(backups) > b.HistorySize {
		removeList = backups[b.HistorySize:]
	}

	log.Println("backups:")
	pprint(backups)

	log.Println("backups to remove: ")
	pprint(removeList)

	if len(removeList) > 0 {
		if err := b.Remove(ctx, removeList); err != nil {
			return err
		}
	}

	return nil
}

func (b *Backup) List(ctx context.Context) error {
	backups, err := b.listObjects(ctx, fmt.Sprintf("%s/%s-", b.Name, b.Database))
	if err != nil {
		return err
	}

	// sort by date
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	log.Println("backups: ")
	pprint(backups)

	return nil
}
