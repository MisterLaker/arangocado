package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var CmdCleanUp = &cobra.Command{
	Use:   "cleanup",
	Short: "remove old backups",
	Run:   runCleanUp,
}

func init() {
	CmdCleanUp.PersistentFlags().StringVar(&cfg.Name, "name", "arangocado", "backup name")
	CmdCleanUp.PersistentFlags().StringVar(&cfg.ConfigFile, "config", "config.yaml", "config file")
	CmdCleanUp.PersistentFlags().StringSliceVar(&cfg.Storages, "storage", []string{}, "backup storage to remove backup")
}

func runCleanUp(c *cobra.Command, args []string) {
	config, err := loadConfig(cfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	sr, err := newStorageRegistry(config)
	if err != nil {
		log.Fatalln("Unable to create storages", err)
	}

	backup := config.GetBackup(cfg.Name)
	if backup == nil {
		log.Fatalln("Unable to find backup", cfg.Name)
	}

	ctx := context.Background()

	b, err := newBackup(*backup, sr)
	if err != nil {
		log.Fatalln("Unable to init backup", err)
	}

	if err := b.CleanUp(ctx, cfg.Storages...); err != nil {
		log.Fatalln("Unable to remove backups", err)
	}
}
