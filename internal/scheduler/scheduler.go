package scheduler

import (
	"context"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/MisterLaker/arangocado/internal/backups"
)

type Scheduler struct {
	checkIterval time.Duration
	backups      timeList
}

func New(checkIterval time.Duration, backups []*BackupSchedule) *Scheduler {
	if checkIterval == 0 {
		checkIterval = 10 * time.Second
	}

	return &Scheduler{
		checkIterval: checkIterval,
		backups:      timeList(backups),
	}
}

func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.checkIterval)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			s.runBackups(ctx, t)
		}
	}
}

func (s *Scheduler) runBackups(ctx context.Context, t time.Time) {
	items := s.backups.Filter(t)

	for _, b := range items {
		slog.Info("run backup", slog.String("name", b.Name), slog.Time("time", t))

		if err := b.Dump(ctx, &backups.DumpOptions{}); err != nil {
			slog.Error("unable to run backup", slog.String("name", b.Name), slog.Any("err", err))
		}

		if err := b.CleanUp(ctx); err != nil {
			slog.Error("unable to clean up backups", slog.String("name", b.Name), slog.Any("err", err))
		}

		b.SetNextUpdate(t)

		slog.Info("backup scheduled", slog.String("name", b.Name), slog.Time("next_update_at", b.NextUpdateAt))
	}
}

type BackupSchedule struct {
	*backups.Backup
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
