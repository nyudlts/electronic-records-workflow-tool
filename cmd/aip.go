package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(aipCmd)
}

var aipCmd = &cobra.Command{
	Use:   "aip",
	Short: "ADOC AIP commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ADOC AIP tools")
	},
}
