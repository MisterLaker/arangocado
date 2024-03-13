package storages

import (
	"context"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

type mockStorage struct{}

func (s *mockStorage) Upload(ctx context.Context, backup, db, ts, cacheDir string, files []string) error {
	return nil
}

func (s *mockStorage) Download(ctx context.Context, cacheDir string, files []string) error {
	return nil
}

func (s *mockStorage) GetList(ctx context.Context, prefix string) ([]string, error) {
	return nil, nil
}

func (s *mockStorage) Remove(ctx context.Context, files []string) error {
	return nil
}

func (s *mockStorage) CanUpload(t time.Time) bool {
	return true
}

type scheduleTestCase struct {
	Time      time.Time
	CanUpload bool
}

func TestScheduleStorage(t *testing.T) {
	testCases := []struct {
		Schedule   string
		StartTime  time.Time
		NextUpdate time.Time
		Cases      []scheduleTestCase
	}{
		{
			Schedule:   "0 */4 * * *",
			StartTime:  time.Date(2009, time.November, 10, 17, 0, 0, 0, time.UTC),
			NextUpdate: time.Date(2009, time.November, 10, 20, 0, 0, 0, time.UTC),
			Cases: []scheduleTestCase{
				{
					CanUpload: false,
					Time:      time.Date(2009, time.November, 10, 19, 0, 0, 0, time.UTC),
				},
				{
					CanUpload: false,
					Time:      time.Date(2009, time.November, 10, 19, 59, 0, 0, time.UTC),
				},
				{
					CanUpload: true,
					Time:      time.Date(2009, time.November, 10, 20, 0, 0, 1, time.UTC),
				},
				{
					CanUpload: true,
					Time:      time.Date(2009, time.November, 10, 21, 0, 0, 0, time.UTC),
				},
			},
		},
	}

	for i, test := range testCases {
		schedule, err := cron.ParseStandard(test.Schedule)
		if err != nil {
			t.Fatalf("Test '%d' unexpected error: %v", i, err)
		}

		storage := NewScheduleStorage(&mockStorage{}, schedule, test.StartTime)

		if test.NextUpdate != storage.nextUpdateAt {
			t.Errorf("Test '%d' expected: %v \nActual: %v", i, test.NextUpdate, storage.nextUpdateAt)
		}

		for j, tcase := range test.Cases {
			if ok := storage.CanUpload(tcase.Time); tcase.CanUpload != ok {
				t.Errorf("Test '%d-%d' expected: %v \nActual: %v", i, j, tcase.CanUpload, ok)
			}
		}
	}
}
