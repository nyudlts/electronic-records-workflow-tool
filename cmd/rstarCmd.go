package cmd

import (
	"fmt"

	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	rstarSizeCmd.Flags().BoolVarP(&directories, "directories", "d", false, "print directories")
	rstarCmd.AddCommand(rstarSizeCmd)
	rstarPrepPackagesCmd.AddCommand(rstarPrepPackagesLogCmd)
	rstarPrepCmd.AddCommand(rstarPrepPackagesCmd)
	rstarPrepSingleCmd.Flags().StringVarP(&aipLoc, "aip-location", "a", "", "path to the AIP package")
	rstarPrepCmd.AddCommand(rstarPrepSingleCmd)
	rstarCmd.AddCommand(rstarPrepCmd)
	rstarValidateCmd.Flags().BoolVarP(&fullValidation, "full", "f", false, "perform full validation")
	rstarValidateCmd.AddCommand(rstarValidateLogCmd)
	rstarCmd.AddCommand(rstarValidateCmd)
	rstarTransferCmd.AddCommand(rstarTransferLogCmd)
	rstarCmd.AddCommand(rstarTransferCmd)
	rstarCmd.AddCommand(rstarCleanCmd)
	rstarAIPFileCmd.AddCommand(rstarAIPFileLogCmd)
	rstarCmd.AddCommand(rstarAIPFileCmd)
	rootCmd.AddCommand(rstarCmd)
}

var rstarCmd = &cobra.Command{
	Use:   "rstar",
	Short: "ewt aip commands",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var rstarSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "ewt aip size commands",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.PrintRStarPackageSize(directories); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarPrepCmd = &cobra.Command{
	Use:   "prep",
	Short: "ewt aip prep commands",
}

var rstarPrepPackagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "ewt aip prep packages commands",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.PrepareRStarPackages(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarPrepPackagesLogCmd = &cobra.Command{
	Use:   "log",
	Short: "display rstar prep packages log command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.RSTAR_PREP_PACKAGES); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarPrepSingleCmd = &cobra.Command{
	Use:   "single",
	Short: "ewt aip prep single package commands",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.PrepareSinglePackage(aipLoc); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarPrepSingleLogCmd = &cobra.Command{
	Use:   "log",
	Short: "display rstar prep single package log command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.RSTAR_PREP_PACKAGE); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "ewt rstar validate command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ValidateRStarPackages(fullValidation); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarValidateLogCmd = &cobra.Command{
	Use:   "log",
	Short: "display rstar validate log command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.RSTAR_VALIDATE); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "ewt rstar transfer command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.TransferRStarPackages(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarTransferLogCmd = &cobra.Command{
	Use:   "log",
	Short: "display rstar transfer log command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.RSTAR_TRANSFER); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "remove all content from aips directory",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.CleanAIPDirectory(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var rstarAIPFileCmd = &cobra.Command{
	Use:   "aipfile",
	Short: "ewt aipfile commands",
	Run:   func(cmd *cobra.Command, args []string) {},
}

var rstarAIPFileLogCmd = &cobra.Command{
	Use:   "log",
	Short: "ewt aipfile commands",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.AIP_FILE); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}
