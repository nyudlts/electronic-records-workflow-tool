package cmd

import (
	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	rstarSizeCmd.Flags().BoolVarP(&directories, "directories", "d", false, "print directories")
	rstarCmd.AddCommand(rstarSizeCmd)
	rstarPrepCmd.AddCommand(rstarPrepPackagesCmd)
	rstarPrepSingleCmd.Flags().StringVarP(&aipLoc, "path", "p", "", "path to the AIP package")
	rstarPrepCmd.AddCommand(rstarPrepSingleCmd)
	rstarCmd.AddCommand(rstarPrepCmd)
	rstarCmd.AddCommand(rstarValidateCmd)
	rstarCmd.AddCommand(rstarTransferCmd)
	rstarCmd.AddCommand(rstarCleanCmd)
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
			panic(err)
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
			panic(err)
		}
	},
}

var rstarPrepSingleCmd = &cobra.Command{
	Use:   "single",
	Short: "ewt aip prep single package commands",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.PrepareSinglePackage(aipLoc); err != nil {
			panic(err)
		}
	},
}

var rstarValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "ewt rstar validate command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ValidateRStarPackages(); err != nil {
			panic(err)
		}
	},
}

var rstarTransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "ewt rstar transfer command",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.TransferRStarPackages(); err != nil {
			panic(err)
		}
	},
}

var rstarCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "remove all content from aips directory",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.CleanAIPDirectory(); err != nil {
			panic(err)
		}
	},
}
