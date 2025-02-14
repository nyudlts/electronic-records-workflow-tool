package cmd

import "github.com/spf13/cobra"

func init() {
	projectCmd.AddCommand(archiveCmd)
}

var archiveCmd = &cobra.Command{
	Use: "archive",
	Run: func(cmd *cobra.Command, args []string) {},
}
