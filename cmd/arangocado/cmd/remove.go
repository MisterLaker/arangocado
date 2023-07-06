package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var CmdRemove = &cobra.Command{
	Use:   "remove",
	Short: "remove backup",
	Run:   runRemove,
}

var removeCfg struct {
	Name       string
	ConfigFile string
	TS         string
}

func init() {
	CmdRemove.PersistentFlags().StringVar(&removeCfg.Name, "n", "arangocado", "backup name")
	CmdRemove.PersistentFlags().StringVar(&removeCfg.TS, "ts", "", "backup timestamp")
	CmdRemove.PersistentFlags().StringVar(&removeCfg.ConfigFile, "config", "config.yaml", "config file")
}

func runRemove(c *cobra.Command, args []string) {
	config, err := loadConfig(removeCfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	m, err := newMinioClient(config.S3)
	if err != nil {
		log.Fatalln("Unable to create S3 client", err)
	}

	backup := config.GetBackup(removeCfg.Name)
	if backup == nil {
		log.Fatalln("Unable to find backup", removeCfg.Name)
	}

	ctx := context.Background()

	b := newBackup(*backup, config.S3, m)

	if err := b.Remove(ctx, []string{removeCfg.TS}); err != nil {
		log.Fatalln("Unable to remove backups", err)
	}
}
