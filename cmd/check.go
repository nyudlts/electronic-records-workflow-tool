package cmd

import "github.com/spf13/cobra"

func init() {
	checkCmd.Flags().StringVar(&sourceLoc, "source", "", "")
	checkCmd.Flags().StringVar(&stagingLoc, "staging", "", "")
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use: "check",
	Run: func(cmd *cobra.Command, args []string) {
		check()
	},
}

func check() {

}
