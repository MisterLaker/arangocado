package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/MisterLaker/arangocado/internal/backup"
)

type Scheduler struct {
	checkIterval time.Duration
	backups      timeList
}

func New(checkIterval time.Duration, backups []*BackupSchedule) *Scheduler {
	return &Scheduler{
		checkIterval: checkIterval,
		backups:      timeList(backups),
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		t := time.Now()

		items := s.backups.Filter(t)

		for _, b := range items {
			log.Println("run backup", b.Name)

			if err := b.Run(ctx); err != nil {
				log.Println("Unable to run backup", err)
			}

			if err := b.CleanUp(ctx); err != nil {
				log.Println("Unable to clean up backups", err)
			}

			b.SetNextUpdate(t)

			log.Println("backup", b.Name, "nextUpdateAt", b.NextUpdateAt)
		}

		time.Sleep(s.checkIterval)
	}
}

type BackupSchedule struct {
	*backup.Backup
	Schedule     cron.Schedule
	NextUpdateAt time.Time
}

func (b *BackupSchedule) SetNextUpdate(t time.Time) {
	b.NextUpdateAt = b.Schedule.Next(t)
}

type timeList []*BackupSchedule

func (ds timeList) Filter(t time.Time) timeList {
	var items timeList

	for _, d := range ds {
		if d.NextUpdateAt.Before(t) {
			items = append(items, d)
		}
	}

	return items
}
