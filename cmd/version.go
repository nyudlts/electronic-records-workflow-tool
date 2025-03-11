package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "print the version of erwt",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("adoc %s\n", version)
	},
}
