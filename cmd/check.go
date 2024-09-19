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
	/*
		The check command will ensure:
			1. that there is a valid work-order
			2. that there is a valid transfer-info.txt
			3. that there is a ER directory in the transfer for each row in a work order
			4. that there are no ER directories in the transfer that are not listed in the work order
	*/
}
