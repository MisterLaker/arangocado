package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var CmdAuto = &cobra.Command{
	Use:   "auto",
	Short: "run auto backup",
	Run:   runAuto,
}

var cfg struct {
	Name       string
	ConfigFile string
}

func init() {
	CmdAuto.PersistentFlags().StringVar(&cfg.ConfigFile, "config", "config.yaml", "config file")
}

func runAuto(c *cobra.Command, args []string) {
	config, err := loadConfig(cfg.ConfigFile)
	if err != nil {
		log.Fatalln("Unable to load config", err)
	}

	sr, err := newStorageRegistry(config)
	if err != nil {
		log.Fatalln("Unable to create storages", err)
	}

	s, err := newScheduler(config, sr)
	if err != nil {
		log.Fatalln("Unable to create scheduler", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go s.Run(ctx)

	wait()
	cancel()
}
