package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/MisterLaker/arangocado/internal/backup"
)

var CmdRestore = &cobra.Command{
	Use:   "restore",
	Short: "restore backup",
	Run:   runRestore,
}

var restoreCfg struct {
	Name       string
	ConfigFile string
	Database   string
	TS         string
}

func init() {
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.ConfigFile, "config", "config.yaml", "config file")
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.Name, "n", "arangocado", "backup name")
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.TS, "ts", "", "backup timestamp")
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.Database, "db", "", "new database to restore backup")
}

func runRestore(c *cobra.Command, args []string) {
	config, err := loadConfig(restoreCfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	sr, err := newStorageRegistry(config)
	if err != nil {
		log.Fatalln("Unable to create storages", err)
	}

	bc := config.GetBackup(restoreCfg.Name)
	if bc == nil {
		log.Fatalln("Unable to find backup", restoreCfg.Name)
	}

	ctx := context.Background()

	b, err := newBackup(*bc, sr)
	if err != nil {
		log.Fatalln("Unable to init backup", err)
	}

	o := &backup.RestoreOptions{
		Key:      restoreCfg.TS,
		Database: restoreCfg.Database,
	}

	if err := b.Restore(ctx, o); err != nil {
		log.Fatalln("Unable to restore backup", err)
	}
}
