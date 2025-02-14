package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(aspaceCmd)
}

var aspaceCmd = &cobra.Command{
	Use:   "aspace",
	Short: "aspace commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aspace called")
	},
}
