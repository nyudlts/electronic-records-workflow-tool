package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(aipCmd)
}

var aipCmd = &cobra.Command{
	Use:   "aip",
	Short: "erwt aip commands",
	Run:   func(cmd *cobra.Command, args []string) {},
}
