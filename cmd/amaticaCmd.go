package cmd

import (
	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	amaticaSizeCmd.Flags().BoolVarP(&directories, "directories", "d", false, "print directories")
	amaticaCmd.AddCommand(amaticaSizeCmd)
	amaticaPrepCmd.Flags().IntVar(&numWorkers, "workers", 1, "number of worker threads to process SIPs")
	amaticaCmd.AddCommand(amaticaPrepCmd)
	amaticaClearCmd.Flags().BoolVarP(&ingests, "ingests", "i", false, "")
	amaticaClearCmd.Flags().BoolVarP(&transfers, "transfers", "t", false, "")
	amaticaCmd.AddCommand(amaticaClearCmd)
	amaticaTransferCmd.Flags().StringVarP(&amaticaConfigLoc, "config", "c", "", "path to Archivematica config file")
	amaticaTransferCmd.Flags().IntVar(&pollTime, "poll", 15, "polling time, in seconds, between calls to Archivematica api to check status")
	amaticaCmd.AddCommand(amaticaTransferCmd)
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

var amaticaClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear Archivematica transfers and ingests",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.AmaticaClear(transfers, ingests); err != nil {
			panic(err)
		}
	},
}

var amaticaTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer SIP to Archivematica",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.TransferToArchivematica(pollTime, amaticaConfigLoc); err != nil {
			panic(err)
		}
	},
}
