package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/MisterLaker/arangocado/cmd/arangocado/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "arangocado",
	Short: "arangodb backup tool",
}

func init() {
	rootCmd.AddCommand(cmd.CmdAuto)
	rootCmd.AddCommand(cmd.CmdBackup)
	rootCmd.AddCommand(cmd.CmdCleaUp)
	rootCmd.AddCommand(cmd.CmdList)
	rootCmd.AddCommand(cmd.CmdRemove)
	rootCmd.AddCommand(cmd.CmdRestore)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
