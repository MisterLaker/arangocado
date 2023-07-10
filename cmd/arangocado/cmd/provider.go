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
)

func newMinioClient(config S3) (*minio.Client, error) {
	return minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
}

func newBackup(config Backup, s3 S3, m *minio.Client) *backup.Backup {
	name := "arangocado"
	if config.Name != "" {
		name = config.Name
	}

	cacheDir := "/tmp"
	if config.CacheDir != "" {
		cacheDir = config.CacheDir
	}

	workers := runtime.NumCPU()
	if s3.Workers > 0 {
		workers = s3.Workers
	}

	return &backup.Backup{
		Name:        name,
		Host:        config.Host,
		User:        config.User,
		Password:    config.Password,
		Database:    config.Database,
		Collections: config.Collections,
		CacheDir:    cacheDir,
		HistorySize: config.HistorySize,
		Workers:     workers,
		Bucket:      s3.Bucket,
		Minio:       m,
	}
}

func newBackupSchedule(config Scheduler, s3 S3, m *minio.Client) (*scheduler.BackupSchedule, error) {
	schedule, err := cron.ParseStandard(config.Schedule)
	if err != nil {
		return nil, fmt.Errorf("schedule: %s - %w", config.Schedule, err)
	}

	config.Backup.Name = config.Name

	return &scheduler.BackupSchedule{
		Schedule: schedule,
		Backup:   newBackup(config.Backup, s3, m),
	}, nil
}

func newScheduler(config *Config, m *minio.Client) (*scheduler.Scheduler, error) {
	var backups []*scheduler.BackupSchedule

	t := time.Now()

	for _, bs := range config.Backups {
		if bs.Schedule == "" {
			continue
		}

		bs.CacheDir = config.CacheDir

		b, err := newBackupSchedule(bs, config.S3, m)
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
