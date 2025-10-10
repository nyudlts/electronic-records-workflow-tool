package cmd

import (
	"fmt"

	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

var (
	envPtr    *string
	configPtr *string
)

func init() {
	aspaceCheckCmd.Flags().StringVarP(&aspaceEnv, "environment", "e", "", "")
	aspaceCheckCmd.Flags().StringVarP(&aspaceConfig, "config", "c", "", "")
	aspaceCheckCmd.AddCommand(aspaceCheckLogCmd)
	aspaceCmd.AddCommand(aspaceCheckCmd)
	aspaceHealthCmd.Flags().StringVarP(&aspaceEnv, "environment", "e", "", "")
	aspaceHealthCmd.Flags().StringVarP(&aspaceConfig, "config", "c", "", "")
	aspaceCmd.AddCommand(aspaceHealthCmd)
	rootCmd.AddCommand(aspaceCmd)
}

var aspaceCmd = &cobra.Command{
	Use:   "aspace",
	Short: "ewt ArchivesSpace commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Help())
	},
}

var aspaceCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Check that DOs exist in ArchivesSpace",
	Run: func(cmd *cobra.Command, args []string) {
		setAspacePointers()
		if err := lib.AspaceCheck(configPtr, envPtr); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var aspaceCheckLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Display the log file from a previous aspace check",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.ASPACE_CHECK); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var aspaceHealthCmd = &cobra.Command{
	Use:   "health",
	Short: "check that the aspace instance is available",
	Run: func(cmd *cobra.Command, args []string) {
		setAspacePointers()
		if err := lib.AspaceHealth(configPtr, envPtr); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

func setAspacePointers() {
	if aspaceEnv != "" {
		envPtr = &aspaceEnv
	} else {
		envPtr = nil
	}

	if aspaceConfig != "" {
		configPtr = &aspaceConfig
	} else {
		configPtr = nil
	}
}
