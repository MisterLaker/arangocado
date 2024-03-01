package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var CmdBackup = &cobra.Command{
	Use:   "backup",
	Short: "create backup dump",
	Run:   runBackup,
}

func init() {
	CmdBackup.PersistentFlags().StringVar(&cfg.Name, "n", "arangocado", "backup name")
	CmdBackup.PersistentFlags().StringVar(&cfg.ConfigFile, "config", "config.yaml", "config file")
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

	backup := config.GetBackup(cfg.Name)
	if backup == nil {
		log.Fatalln("Unable to find backup: ", cfg.Name)
	}

	ctx := context.Background()

	b, err := newBackup(*backup, sr)
	if err != nil {
		log.Fatalln("Unable to init backup", err)
	}

	if err := b.Dump(ctx); err != nil {
		log.Fatalln("Unable to create backup", err)
	}
}
