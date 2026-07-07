package backups

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
)

type Database interface {
	Dump(ctx context.Context, cacheDir string) error
	Restore(ctx context.Context, options *RestoreOptions) error
	GetName() string
}

type ArangoDB struct {
	Host        string
	Port        string
	User        string
	Password    string
	Database    string
	Collections []string
}

func (b *ArangoDB) GetName() string {
	return b.Database
}

func (b *ArangoDB) Dump(ctx context.Context, cacheDir string) error {
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

	cmd := exec.CommandContext(ctx, "arangodump", makeCommandArgs(args)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (b *ArangoDB) Restore(ctx context.Context, options *RestoreOptions) error {
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

	slog.Info("arangorestore", slog.String("database", db), slog.String("cache_dir", options.CacheDir))

	cmd := exec.CommandContext(ctx, "arangorestore", makeCommandArgs(args)...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func makeCommandArgs(args map[string]any) []string {
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
