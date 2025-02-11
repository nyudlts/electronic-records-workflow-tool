package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(sipCmd)
}

var sipCmd = &cobra.Command{
	Use: "sip",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("adoc sip tools")
	},
}
