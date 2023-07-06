package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var CmdList = &cobra.Command{
	Use:   "list",
	Short: "list of arangodb backups",
	Run:   runList,
}

func init() {
	CmdList.PersistentFlags().StringVar(&cfg.Name, "n", "", "backup name")
	CmdList.PersistentFlags().StringVar(&cfg.ConfigFile, "config", "config.yaml", "config file")
}

func runList(c *cobra.Command, args []string) {
	config, err := loadConfig(cfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	m, err := newMinioClient(config.S3)
	if err != nil {
		log.Fatalln("Unable to create S3 client", err)
	}

	var backups []Backup

	if cfg.Name != "" {
		bc := config.GetBackup(cfg.Name)
		if bc == nil {
			log.Fatalln("Unable to find backup", cfg.Name)
		}

		backups = append(backups, *bc)
	}

	if len(backups) == 0 {
		for _, bc := range config.Backups {
			backups = append(backups, bc.Backup)
		}
	}

	ctx := context.Background()

	for _, bc := range backups {
		b := newBackup(bc, config.S3, m)

		log.Printf("name: %s\n", b.Name)

		if err := b.List(ctx); err != nil {
			log.Fatalln("Unable to get list if backups", err)
		}
	}
}
