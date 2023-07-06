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
	CmdCleaUp.PersistentFlags().StringVar(&cfg.ConfigFile, "config", "config.yaml", "config file")
}

func runCleanUp(c *cobra.Command, args []string) {
	config, err := loadConfig(cfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	m, err := newMinioClient(config.S3)
	if err != nil {
		log.Fatalln("Unable to create S3 client", err)
	}

	ctx := context.Background()

	b := newBackup(config, m)

	if err := b.CleanUp(ctx); err != nil {
		log.Fatalln("Unable to remove backups", err)
	}
}
