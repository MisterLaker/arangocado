package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/robfig/cron/v3"

	"github.com/MisterLaker/arangocado/internal/backup"
	"github.com/MisterLaker/arangocado/internal/scheduler"
	"github.com/MisterLaker/arangocado/internal/storages"
)

func newMinioClient(config S3) (*minio.Client, error) {
	return minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
}

func newStorage(config S3) (storages.Storage, error) {
	m, err := newMinioClient(config)
	if err != nil {
		return nil, err
	}

	workers := config.Workers

	if workers == 0 {
		workers = runtime.NumCPU()
	}

	storage := &storages.S3{
		Minio:  m,
		Bucket: config.Bucket,
	}

	return &storages.ChunkStorage{Storage: storage, Workers: workers}, nil
}

func newStorageRegistry(config *Config) (storages.Registry, error) {
	registry := make(storages.Registry)

	for key, cfg := range config.Storages {
		s, err := newStorage(cfg)
		if err != nil {
			return nil, err
		}

		registry[key] = s
	}

	return registry, nil
}

func newBackup(config Backup, storage storages.Registry) (*backup.Backup, error) {
	if len(config.Storages) == 0 {
		return nil, fmt.Errorf("storages is not defined: %s", config.Name)
	}

	for _, s := range config.Storages {
		if _, err := storage.Get(s); err != nil {
			return nil, fmt.Errorf("backup: %s, %w: %s", config.Name, err, s)
		}
	}

	return &backup.Backup{
		Name:        config.Name,
		Host:        config.Host,
		User:        config.User,
		Password:    config.Password,
		Database:    config.Database,
		Collections: config.Collections,
		CacheDir:    config.CacheDir,
		HistorySize: config.HistorySize,
		Storages:    config.Storages,
		Registry:    storage,
	}, nil
}

func newBackupSchedule(config Scheduler, storage storages.Registry) (*scheduler.BackupSchedule, error) {
	schedule, err := cron.ParseStandard(config.Schedule)
	if err != nil {
		return nil, fmt.Errorf("schedule: %s - %w", config.Schedule, err)
	}

	backup, err := newBackup(config.Backup, storage)
	if err != nil {
		return nil, err
	}

	return &scheduler.BackupSchedule{
		Schedule: schedule,
		Backup:   backup,
	}, nil
}

func newScheduler(config *Config, storage storages.Registry) (*scheduler.Scheduler, error) {
	var backups []*scheduler.BackupSchedule

	t := time.Now()

	for _, bs := range config.Backups {
		if bs.Schedule == "" {
			continue
		}

		b, err := newBackupSchedule(bs, storage)
		if err != nil {
			return nil, err
		}

		if !bs.TriggerImmediately {
			b.SetNextUpdate(t)
		}

		backups = append(backups, b)
	}

	return scheduler.New(config.CheckInterval, backups), nil
}

func wait() {
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutdown ...")
}
