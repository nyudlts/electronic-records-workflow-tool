package cmd

import (
	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	// Add your commands here
	amaticaSizeCmd.Flags().BoolVarP(&directories, "directories", "d", false, "print directories")
	amaticaCmd.AddCommand(amaticaSizeCmd)
	rootCmd.AddCommand(amaticaCmd)
}

var amaticaCmd = &cobra.Command{
	Use:   "amatica",
	Short: "ewt Archivematica commands",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var amaticaSizeCmd = &cobra.Command{
	Use: "size",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.PrintXferPackageSize(directories); err != nil {
			panic(err)
		}
	},
}
