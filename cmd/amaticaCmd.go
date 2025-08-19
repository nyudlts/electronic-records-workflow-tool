package cmd

import (
	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	// Add your commands here
	amaticaSizeCmd.Flags().BoolVarP(&directories, "directories", "d", false, "print directories")
	amaticaCmd.AddCommand(amaticaSizeCmd)
	amaticaPrepCmd.Flags().IntVar(&numWorkers, "workers", 1, "number of worker threads to process SIPs")
	amaticaCmd.AddCommand(amaticaPrepCmd)
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

var amaticaPrepCmd = &cobra.Command{
	Use:   "prep",
	Short: "Prepare SIP package for transfer to Archivematica",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.PrepAmatica(numWorkers); err != nil {
			panic(err)
		}
	},
}
