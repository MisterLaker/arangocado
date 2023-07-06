package backup

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"

	"github.com/minio/minio-go/v7"
	"golang.org/x/sync/errgroup"
)

type Backup struct {
	Name            string
	Host            string
	Port            string
	User            string
	Password        string
	Database        string
	Collections     []string
	Directory       string
	Bucket          string
	Workers         int
	KeepLastBackups int

	Minio *minio.Client
}

func (b *Backup) Create(ctx context.Context) error {
	args := map[string]any{
		"server.endpoint":  b.Host,
		"server.username":  b.User,
		"server.password":  b.Password,
		"server.database":  b.Database,
		"collection":       b.Collections,
		"output-directory": b.Directory,
		"overwrite":        true,
	}

	cmd := exec.CommandContext(ctx, "arangodump", makeCmdArgs(args)...)

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

func (b *Backup) UploadFiles(ctx context.Context, ts string, files []string) error {
	for _, name := range files {
		objectName := fmt.Sprintf("%s/%s-%s/%s", b.Name, b.Database, ts, name)
		path := filepath.Join(b.Directory, name)

		fmt.Println("file: ", objectName, ", path: ", path)

		_, err := b.Minio.FPutObject(ctx, b.Bucket, objectName, path, minio.PutObjectOptions{})
		if err != nil {
			fmt.Println("upload error: ", err)

			continue
		}

		fmt.Println("Successfully uploaded file: ", path)
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

func (b *Backup) Upload(ctx context.Context) error {
	files, err := listFiles(b.Directory)
	if err != nil {
		return err
	}

	ts := time.Now().Format("20060102T1504")

	ln := len(files)
	chunkSize := (ln + b.Workers - 1) / b.Workers

	fmt.Println("upload files ", "files: ", ln, "chunk_size: ", chunkSize)

	g, ctx := errgroup.WithContext(ctx)

	for i := 0; i < ln; i += chunkSize {
		i := i
		end := i + chunkSize

		if end > ln {
			end = ln
		}

		g.Go(func() error {
			return b.UploadFiles(ctx, ts, files[i:end])
		})
	}

	return g.Wait()
}

func (b *Backup) remove(ctx context.Context, keys []string) error {
	for _, k := range keys {
		objects, err := b.listObjects(ctx, k)
		if err != nil {
			return err
		}

		for _, name := range objects {
			fmt.Println("removing: ", name)

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

func (b *Backup) CleanUp(ctx context.Context) error {
	backups, err := b.listObjects(ctx, fmt.Sprintf("%s/%s-", b.Name, b.Database))
	if err != nil {
		return err
	}

	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	var removeList []string

	if len(backups) > b.KeepLastBackups {
		removeList = backups[b.KeepLastBackups:]
	}

	fmt.Println("backups: ", backups)
	fmt.Println("backups to remove: ", removeList)

	if len(removeList) > 0 {
		if err := b.remove(ctx, removeList); err != nil {
			return err
		}
	}

	return nil
}
