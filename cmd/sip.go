package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sipCmd)
}

var sipCmd = &cobra.Command{
	Use:   "sip",
	Short: "erwt sip commands",
	Run:   func(cmd *cobra.Command, args []string) {},
}
