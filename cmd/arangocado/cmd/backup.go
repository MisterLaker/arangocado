package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/MisterLaker/arangocado/internal/backups"
)

var CmdBackup = &cobra.Command{
	Use:   "backup",
	Short: "create backup dump",
	Run:   runBackup,
}

func init() {
	CmdBackup.PersistentFlags().StringVar(&cfg.Name, "name", "arangocado", "backup name")
	CmdBackup.PersistentFlags().StringVar(&cfg.ConfigFile, "config", "config.yaml", "config file")
	CmdBackup.PersistentFlags().StringSliceVar(&cfg.Storages, "storage", []string{}, "backup storage to remove backup")
}

func runBackup(c *cobra.Command, args []string) {
	config, err := loadConfig(cfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	sr, err := newStorageRegistry(config)
	if err != nil {
		log.Fatalln("Unable to create storages", err)
	}

	backupConfig := config.GetBackup(cfg.Name)
	if backupConfig == nil {
		log.Fatalln("Unable to find backup: ", cfg.Name)
	}

	ctx := context.Background()

	backup, err := newBackup(*backupConfig, sr)
	if err != nil {
		log.Fatalln("Unable to init backup", err)
	}

	options := &backups.DumpOptions{
		SkipSchedule: true,
		Storages:     cfg.Storages,
	}

	if err := backup.Dump(ctx, options); err != nil {
		log.Fatalln("Unable to create backup", err)
	}
}
