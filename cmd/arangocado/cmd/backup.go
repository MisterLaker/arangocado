package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var CmdBackup = &cobra.Command{
	Use:   "backup",
	Short: "run arangodb dump",
	Run:   runBackup,
}

var cfg struct {
	ConfigFile string
}

func init() {
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

	ctx := context.Background()

	b := newBackup(config, m)

	if err := b.RemoveCache(); err != nil {
		log.Fatalln("Unable to remove cache", err)
	}

	if err := b.Create(ctx); err != nil {
		log.Fatalln("Unable to create backup", err)
	}

	if err := b.Upload(ctx); err != nil {
		log.Fatalln("Unable to upload backup", err)
	}
}
