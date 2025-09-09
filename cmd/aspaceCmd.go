package cmd

import (
	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	aspaceCheckCmd.Flags().StringVarP(&aspaceEnv, "environment", "e", "", "")
	aspaceCmd.AddCommand(aspaceCheckCmd)
	aspaceHealthCmd.Flags().StringVarP(&aspaceEnv, "environment", "e", "", "")
	aspaceCmd.AddCommand(aspaceHealthCmd)
	rootCmd.AddCommand(aspaceCmd)
}

var aspaceCmd = &cobra.Command{
	Use:   "aspace",
	Short: "ewt ArchivesSpace commands",
}

var envPtr *string

var aspaceCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check that DOs exist in ArchivesSpace",
	Run: func(cmd *cobra.Command, args []string) {

		if aspaceEnv != "" {
			envPtr = &aspaceEnv
		} else {
			envPtr = nil
		}

		if err := lib.AspaceCheck(envPtr); err != nil {
			panic(err)
		}
	},
}

var aspaceHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "check that the aspace instance is available",
	Run: func(cmd *cobra.Command, args []string) {
		if aspaceEnv != "" {
			envPtr = &aspaceEnv
		} else {
			envPtr = nil
		}

		if err := lib.AspaceHealth(envPtr); err != nil {
			panic(err)
		}
	},
}
