package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"

	"github.com/MisterLaker/arangocado/internal/backups"
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
	Storage    string
}

func init() {
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.ConfigFile, "config", "config.yaml", "config file")
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.Name, "n", "arangocado", "backup name")
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.TS, "ts", "", "backup timestamp")
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.Database, "db", "", "new database to restore backup")
	CmdRestore.PersistentFlags().StringVar(&restoreCfg.Storage, "storage", "", "backup storage to restore backup")
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

	backupConfig := config.GetBackup(restoreCfg.Name)
	if backupConfig == nil {
		log.Fatalln("Unable to find backup", restoreCfg.Name)
	}

	ctx := context.Background()

	backup, err := newBackup(*backupConfig, sr)
	if err != nil {
		log.Fatalln("Unable to init backup", err)
	}

	options := &backups.RestoreOptions{
		Key:      restoreCfg.TS,
		Database: restoreCfg.Database,
		Storage:  restoreCfg.Storage,
	}

	if err := backup.Restore(ctx, options); err != nil {
		log.Fatalln("Unable to restore backup", err)
	}
}
