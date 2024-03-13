package backups

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/MisterLaker/arangocado/internal/storages"
)

type mockDatabase struct {
	dumpFiles     []string
	restoredFiles []string
}

func (d *mockDatabase) Dump(ctx context.Context, cacheDir string) error {
	// check if dir exists
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return err
	}

	file := "dumpdb.data"

	err := os.WriteFile(filepath.Join(cacheDir, file), []byte("some data"), 0666)
	if err != nil {
		return err
	}

	d.dumpFiles = append(d.dumpFiles, file)

	return nil
}

func (d *mockDatabase) Restore(ctx context.Context, options *RestoreOptions) error {
	files, err := listFiles(options.CacheDir)
	if err != nil {
		return err
	}

	d.restoredFiles = files

	return nil
}

func (d *mockDatabase) GetName() string {
	return "db"
}

type mockStorage struct {
	uploadFiles     []string
	downloadedFiles []string
}

func (s *mockStorage) Upload(ctx context.Context, backup, db, ts, cacheDir string, files []string) error {
	s.uploadFiles = files

	return nil
}

func (s *mockStorage) Download(ctx context.Context, cacheDir string, files []string) error {
	for _, f := range files {
		path := filepath.Join(cacheDir, f)
		dir, file := filepath.Split(path)

		// check if dir exists
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		if err := os.WriteFile(path, []byte("some data"), 0666); err != nil {
			return err
		}

		s.downloadedFiles = append(s.downloadedFiles, file)
	}

	return nil
}

func (s *mockStorage) GetList(ctx context.Context, prefix string) ([]string, error) {
	return []string{filepath.Join(prefix, "restoredb.data")}, nil
}

func (s *mockStorage) Remove(ctx context.Context, files []string) error {
	return nil
}

func (s *mockStorage) CanUpload(t time.Time) bool {
	return true
}

func TestBackupDump(t *testing.T) {
	db := &mockDatabase{}
	storage := &mockStorage{}
	testStorage := "testing"

	registry := storages.Registry{
		testStorage: storage,
	}

	backup := &Backup{
		Name:     "testing",
		Database: db,
		CacheDir: "/tmp",
		Registry: registry,
		Storages: []string{testStorage},
	}

	if err := backup.Dump(context.Background(), &DumpOptions{}); err != nil {
		t.Fatalf("Test unexpected error: %v", err)
	}

	if len(db.dumpFiles) == 0 {
		t.Errorf("Test expected non empty file list: \nActual: %v", db.dumpFiles)
	}

	if !reflect.DeepEqual(db.dumpFiles, storage.uploadFiles) {
		t.Errorf("Test expected: %v \nActual: %v", db.dumpFiles, storage.uploadFiles)
	}
}

func TestBackupRestore(t *testing.T) {
	db := &mockDatabase{}
	storage := &mockStorage{}

	testStorage := "testing"

	registry := storages.Registry{
		testStorage: storage,
	}

	backup := &Backup{
		Name:     "testing",
		Database: db,
		CacheDir: "/tmp",
		Registry: registry,
		Storages: []string{testStorage},
	}

	options := &RestoreOptions{
		Storage: testStorage,
		Key:     time.Now().Format("20060102T1504"),
	}

	if err := backup.Restore(context.Background(), options); err != nil {
		t.Fatalf("Test unexpected error: %v", err)
	}

	if len(db.restoredFiles) == 0 {
		t.Errorf("Test expected non empty file list: \nActual: %v", db.restoredFiles)
	}

	if !reflect.DeepEqual(db.restoredFiles, storage.downloadedFiles) {
		t.Errorf("Test expected: %v \nActual: %v", storage.downloadedFiles, db.restoredFiles)
	}
}
