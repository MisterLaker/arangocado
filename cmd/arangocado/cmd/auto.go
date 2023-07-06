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

	m, err := newMinioClient(config.S3)
	if err != nil {
		log.Fatalln("Unable to create S3 client", err)
	}

	s, err := newScheduler(config, m)
	if err != nil {
		log.Fatalln("Unable to create scheduler", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go s.Run(ctx)

	wait()
	cancel()
}
