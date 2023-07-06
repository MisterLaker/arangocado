package cmd

import (
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/MisterLaker/arangocado/internal/backup"
)

func newMinioClient(config S3) (*minio.Client, error) {
	return minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
}

func newBackup(config *Config, m *minio.Client) *backup.Backup {
	return &backup.Backup{
		Host:        config.Arango.Host,
		User:        config.Arango.User,
		Password:    config.Arango.Password,
		Database:    config.Arango.Database,
		Collections: config.Arango.Collections,
		Directory:   config.Arango.Directory,
		Workers:     config.S3.Workers,
		Bucket:      config.S3.Bucket,
		Minio:       m,
	}
}
