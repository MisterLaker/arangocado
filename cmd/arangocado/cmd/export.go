package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var CmdExport = &cobra.Command{
	Use:   "export",
	Short: "run arangodb dump",
	Run:   runExport,
}

var exportCfg struct {
	ConfigFile string
}

func init() {
	CmdExport.PersistentFlags().StringVar(&exportCfg.ConfigFile, "config", "config.yaml", "config file")
}

func runExport(c *cobra.Command, args []string) {
	config, err := loadConfig(exportCfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	m, err := newMinioClient(config.S3)
	if err != nil {
		log.Fatalln("Unable to create S3 client", err)
	}

	ctx := context.Background()

	b := newBackup(config, m)

	if err := b.Create(ctx); err != nil {
		log.Fatalln("Unable to create backup", err)
	}

	if err := b.Upload(ctx); err != nil {
		log.Fatalln("Unable to upload backup", err)
	}
}
