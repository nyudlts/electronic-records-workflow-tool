package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(aspaceCmd)
}

var aspaceCmd = &cobra.Command{
	Use:   "aspace",
	Short: "ADOC ArchivesSpace commands",
}
