package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// Add your commands here
	rootCmd.AddCommand(amaticaCmd)
}

var amaticaCmd = &cobra.Command{
	Use:   "amatica",
	Short: "ADOC Archivematica commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ADOC Archivematica commands")
	},
}
