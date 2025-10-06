package cmd

import (
	"fmt"

	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	sourceSizeCmd.Flags().BoolVarP(&directories, "directory", "d", false, "Print size info for each directory")
	sourceCmd.AddCommand(sourceSizeCmd)
	sourceXferCmd.AddCommand(sourceXferLogCmd)
	sourceCmd.AddCommand(sourceXferCmd)
	rootCmd.AddCommand(sourceCmd)
}

var sourceCmd = &cobra.Command{
	Use:   "source",
	Short: "ewt source commands",
}

var sourceXferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer source files to the SIP directory",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.TransferSource(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var sourceXferLogCmd = &cobra.Command{
	Use:   "log",
	Short: "View transfer log",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.SOURCE_TRANSFER); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var sourceSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Get size of source directory",
	Run: func(cmd *cobra.Command, args []string) {

		//print the total size of source directory
		if err := lib.PrintSourcePackageSize(directories); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}
