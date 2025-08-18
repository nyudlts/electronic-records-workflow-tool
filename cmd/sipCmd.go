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
	rootCmd.AddCommand(sipCmd)

}

var sipCmd = &cobra.Command{
	Use:   "sip",
	Short: "ewt sip commands",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sip subcommand executed")
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

var sipScanCmd = &cobra.Command{
	Use: "scan",
	Run: func(cmd *cobra.Command, args []string) {},
}

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
