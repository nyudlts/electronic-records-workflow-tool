package cmd

import (
	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	aspaceCmd.AddCommand(aspaceCheckCmd)
	rootCmd.AddCommand(aspaceCmd)
}

var aspaceCmd = &cobra.Command{
	Use:   "aspace",
	Short: "ewt ArchivesSpace commands",
}

var aspaceCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check that DOs exist in ArchivesSpace",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.AspaceCheck(); err != nil {
			panic(err)
		}
	},
}
