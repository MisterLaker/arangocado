package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var CmdCleaUp = &cobra.Command{
	Use:   "cleanup",
	Short: "remove old backups",
	Run:   runCleanUp,
}

func init() {
	CmdCleaUp.PersistentFlags().StringVar(&cfg.Name, "n", "arangocado", "backup name")
	CmdCleaUp.PersistentFlags().StringVar(&cfg.ConfigFile, "config", "config.yaml", "config file")
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

	if err := b.CleanUp(ctx); err != nil {
		log.Fatalln("Unable to remove backups", err)
	}
}
