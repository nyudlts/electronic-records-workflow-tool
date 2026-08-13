package cmd

import (
	"fmt"

	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {

	sipGenXferCmd.Flags().StringVarP(&profile, "profile", "p", "", "profile initials")
	sipGenCmd.AddCommand(sipGenXferCmd)
	sipCmd.AddCommand(sipGenCmd)
	sipValidateCmd.AddCommand(sipValidateLogCmd)
	sipCmd.AddCommand(sipValidateCmd)
	sipScanAVCmd.AddCommand(sipScanAVLogCmd)
	sipScanCmd.AddCommand(sipScanAVCmd)
	sipScanCleanCmd.AddCommand(sipScanCleanLogCmd)
	sipScanCmd.AddCommand(sipScanCleanCmd)
	sipScanCmd.AddCommand(sipScanDetoxCmd)
	sipScanCmd.AddCommand(sipScanExtensionsCmd)
	sipCmd.AddCommand(sipScanCmd)
	sipSizeCmd.Flags().BoolVarP(&directories, "directories", "d", false, "print directories")
	rootCmd.AddCommand(sipCmd)

}

var sipCmd = &cobra.Command{
	Use:   "sip",
	Short: "ewt sip commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Help())
	},
}

var sipSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Get size of sip directory",
	Run: func(cmd *cobra.Command, args []string) {

		//print the total size of source directory
		if err := lib.PrintSIPPackageSize(directories); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var sipValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate a sip is ready for transfer to Archivematica",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ValidateSIP(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var sipValidateLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Read log of validation results",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.SIP_VALIDATE); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

// sip scan commands
var sipScanCmd = &cobra.Command{
	Use: "scan",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Help())
	},
}

var sipScanAVCmd = &cobra.Command{
	Use: "av",
	Run: func(cmd *cobra.Command, args []string) {
		scanErrors, err := lib.ScanAV()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if len(scanErrors) > 0 {
			fmt.Printf(" *  %d error(s) encountered:\n", len(scanErrors))
			for _, e := range scanErrors {
				fmt.Printf("    %s", e)
			}
		}
	},
}

var sipScanAVLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Read log of AV scan results",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.SIP_SCAN_AV); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var sipScanCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "remove .DS_Store and Thumbs.db files from SIP",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.CleanSip(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var sipScanCleanLogCmd = &cobra.Command{
	Use:   "log",
	Short: "Read log of removed .DS_Store and Thumbs.db files from SIP",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ReadLog(lib.SIP_SCAN_CLEAN); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var sipScanDetoxCmd = &cobra.Command{
	Use:   "detox",
	Short: "sub command to run detox on file and directory names",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ScanDetox(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

// sip generator
var sipGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "sub command for generate commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cmd.Help())
	},
}

var sipGenXferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "generate a transfer-info.txt in SIP MD dir",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.GenerateTransferInfo(profile); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}

var sipScanExtensionsCmd = &cobra.Command{
	Use:   "extensions",
	Short: "remediate double file extensions in a sip",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ScanDoubleExtensions(); err != nil {
			fmt.Printf("  * error encountered: %v\n", err)
		}
	},
}
