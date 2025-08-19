package cmd

import (
	"fmt"

	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	sipCmd.AddCommand(sipCleanCmd)
	sipGenXferCmd.Flags().StringVarP(&profile, "profile", "p", "", "profile initials")
	sipGenCmd.AddCommand(sipGenXferCmd)
	sipCmd.AddCommand(sipGenCmd)
	sipCmd.AddCommand(sipValidateCmd)
	sipScanCmd.AddCommand(sipScanAVCmd)
	sipCmd.AddCommand(sipScanCmd)
	sipSizeCmd.Flags().BoolVarP(&directories, "directories", "d", false, "print directories")
	sipCmd.AddCommand(sipSizeCmd)
	rootCmd.AddCommand(sipCmd)

}

var sipCmd = &cobra.Command{
	Use:   "sip",
	Short: "ewt sip commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sip subcommand executed")
	},
}

var sipSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Get size of sip directory",
	Run: func(cmd *cobra.Command, args []string) {

		//print the total size of source directory
		if err := lib.PrintSIPPackageSize(directories); err != nil {
			panic(err)
		}
	},
}

var sipCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "remove .DS_Store and Thumbs.db files from SIP",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.CleanSip(); err != nil {
			panic(err)
		}
	},
}

var sipValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "validate a sip is ready for transfer to Archivematica",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ValidateSIP(); err != nil {
			panic(err)
		}
	},
}

// sip scan commands
var sipScanCmd = &cobra.Command{
	Use: "scan",
	Run: func(cmd *cobra.Command, args []string) {},
}

var sipScanAVCmd = &cobra.Command{
	Use: "av",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.ScanAV(); err != nil {
			panic(err)
		}
	},
}

// sip generator
var sipGenCmd = &cobra.Command{
	Use:   "gen",
	Short: "sub command for generate commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sip gen command executed")
	},
}

var sipGenXferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "generate a transfer-info.txt in SIP MD dir",
	Run: func(cmd *cobra.Command, args []string) {
		if err := lib.GenerateTransferInfo(profile); err != nil {
			panic(err)
		}
	},
}
