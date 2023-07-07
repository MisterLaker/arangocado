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

	m, err := newMinioClient(config.S3)
	if err != nil {
		log.Fatalln("Unable to create S3 client", err)
	}

	backup := config.GetBackup(cfg.Name)
	if backup == nil {
		log.Fatalln("Unable to find backup", cfg.Name)
	}

	ctx := context.Background()

	b := newBackup(*backup, config.S3, m)

	if err := b.RemoveCache(); err != nil {
		log.Fatalln("Unable to remove cache", err)
	}

	if err := b.Arangodump(ctx); err != nil {
		log.Fatalln("Unable to create backup", err)
	}

	if err := b.Upload(ctx); err != nil {
		log.Fatalln("Unable to upload backup", err)
	}
}
