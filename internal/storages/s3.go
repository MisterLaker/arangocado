package storages

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
)

type S3 struct {
	Bucket string
	Minio  *minio.Client
}

func (s *S3) Upload(ctx context.Context, backup, db, key, cacheDir string, files []string) error {
	for _, name := range files {
		path := filepath.Join(cacheDir, name)
		objectName := fmt.Sprintf("%s/%s-%s/%s", backup, db, key, name)

		log.Println("file: ", path)
		log.Println("path: ", objectName)

		_, err := s.Minio.FPutObject(ctx, s.Bucket, objectName, path, minio.PutObjectOptions{})
		if err != nil {
			return fmt.Errorf("s3: %w", err)
		}

		log.Println("Successfully uploaded file: ", path)
	}

	return nil
}

func getLocalPath(p string) string {
	return strings.Replace(p, "/", "_", 1)
}

func (s *S3) Download(ctx context.Context, cacheDir string, files []string) error {
	for _, name := range files {
		localPath := getLocalPath(name)

		path := fmt.Sprintf("%s/%s", cacheDir, localPath)

		log.Println("file: ", path)
		log.Println("path: ", name)

		err := s.Minio.FGetObject(ctx, s.Bucket, name, path, minio.GetObjectOptions{})
		if err != nil {
			return fmt.Errorf("s3: %w", err)
		}

		log.Println("Successfully downloaded file: ", path)
	}

	return nil
}

func (s *S3) GetList(ctx context.Context, prefix string) ([]string, error) {
	objectCh := s.Minio.ListObjects(ctx, s.Bucket, minio.ListObjectsOptions{
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

func (s *S3) Remove(ctx context.Context, files []string) error {
	for _, name := range files {
		if err := s.Minio.RemoveObject(ctx, s.Bucket, name, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("s3: %w", err)
		}
	}

	return nil
}
