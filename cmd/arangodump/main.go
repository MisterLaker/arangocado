package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/MisterLaker/arangodump/cmd/arangodump/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "arangodump",
	Short: "arangodb backup tool",
}

func init() {
	rootCmd.AddCommand(cmd.CmdExport)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
