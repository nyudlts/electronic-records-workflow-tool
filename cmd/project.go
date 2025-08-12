package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(projectCmd)
}

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "ewt project commands",
	Run:   func(cmd *cobra.Command, args []string) {},
}
