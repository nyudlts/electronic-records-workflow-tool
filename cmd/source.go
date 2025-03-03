package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(sourceCmd)
}

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "ADOC source commands",
}
