package cmd

import (
	"fmt"

	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	amaticaSizeCmd.Flags().BoolVarP(&directories, "directories", "d", false, "print directories")
	amaticaCmd.AddCommand(amaticaSizeCmd)
	amaticaPrepCmd.Flags().IntVar(&numWorkers, "workers", 1, "number of worker threads to process SIPs")
	amaticaPrepCmd.AddCommand(amaticaPrepLogCmd)
	amaticaCmd.AddCommand(amaticaPrepCmd)
	amaticaClearCmd.AddCommand(amaticaClearLogCmd)
	amaticaClearCmd.Flags().BoolVarP(&ingests, "ingests", "i", false, "")
	amaticaClearCmd.Flags().BoolVarP(&transfers, "transfers", "t", false, "")
	amaticaCmd.AddCommand(amaticaClearCmd)
	amaticaTransferCmd.AddCommand(amaticaTransferLogCmd)
	amaticaTransferCmd.Flags().StringVarP(&amaticaConfigLoc, "config", "c", "", "path to Archivematica config file")
	amaticaTransferCmd.Flags().IntVar(&pollTime, "poll", 15, "polling time, in seconds, between calls to Archivematica api to check status")
	amaticaCmd.AddCommand(amaticaTransferCmd)
	amaticaCmd.AddCommand(amaticaCountCmd)
	amaticaGenCmd.Flags().StringVarP(&mcpProfile, "profile", "p", "", "")
	amaticaCmd.AddCommand(amaticaGenCmd)
	rootCmd.AddCommand(amaticaCmd)
}

var amaticaCmd = &cobra.Command{
	Use:   "amatica",
	Short: "ewt Archivematica commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Help())
	},
}

var amaticaSizeCmd = &cobra.Command{
	Use: "size",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.PrintXferPackageSize(directories); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var amaticaPrepCmd = &cobra.Command{
	Use:   "prep",
	Short: "Prepare SIP package for transfer to Archivematica",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.PrepAmatica(numWorkers); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var amaticaPrepLogCmd = &cobra.Command{
	Use:   "log",
	Short: "view Archivematica prep log",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.AMATICA_PREP); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var amaticaClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear Archivematica transfers and ingests",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.AmaticaClear(transfers, ingests); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var amaticaClearLogCmd = &cobra.Command{
	Use:   "log",
	Short: "view Archivematica clear log",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.AMATICA_CLEAR); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var amaticaTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer SIP to Archivematica",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.TransferToArchivematica(pollTime, amaticaConfigLoc); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var amaticaTransferLogCmd = &cobra.Command{
	Use:   "log",
	Short: "view Archivematica transfer log",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.AMATICA_TRANSFER); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var amaticaCountCmd = &cobra.Command{
	Use:   "count",
	Short: "Count Archivematica active archivematica transfers",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.CountAmaticaTransfers(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var amaticaGenCmd = &cobra.Command{
	Use: "gen",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.GenerateMCP(mcpProfile); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}
