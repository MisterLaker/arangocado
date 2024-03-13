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
	Storage    string
}

func init() {
	CmdRemove.PersistentFlags().StringVar(&removeCfg.Name, "name", "arangocado", "backup name")
	CmdRemove.PersistentFlags().StringVar(&removeCfg.TS, "ts", "", "backup timestamp")
	CmdRemove.PersistentFlags().StringVar(&removeCfg.ConfigFile, "config", "config.yaml", "config file")
	CmdRemove.PersistentFlags().StringVar(&removeCfg.Storage, "storage", "", "backup storage to remove backup")
}

func runRemove(c *cobra.Command, args []string) {
	config, err := loadConfig(removeCfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	sr, err := newStorageRegistry(config)
	if err != nil {
		log.Fatalln("Unable to create storages", err)
	}

	backup := config.GetBackup(removeCfg.Name)
	if backup == nil {
		log.Fatalln("Unable to find backup", removeCfg.Name)
	}

	ctx := context.Background()

	b, err := newBackup(*backup, sr)
	if err != nil {
		log.Fatalln("Unable to init backup", err)
	}

	if err := b.Remove(ctx, removeCfg.Storage, []string{removeCfg.TS}); err != nil {
		log.Fatalln("Unable to remove backups", err)
	}
}
