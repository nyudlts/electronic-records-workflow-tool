package cmd

import (
	"github.com/nyudlts/electronic-records-workflow-tool/lib"
	"github.com/spf13/cobra"
)

func init() {
	sipCmd.AddCommand(sipCleanCmd)
	rootCmd.AddCommand(sipCmd)

}

var sipCmd = &cobra.Command{
	Use:   "sip",
	Short: "ewt sip commands",
	Run:   func(cmd *cobra.Command, args []string) {},
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
