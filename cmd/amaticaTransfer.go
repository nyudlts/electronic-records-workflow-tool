package cmd

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(xferAmaticaCmd)
}

var xferAmaticaCmd = &cobra.Command{
	Use: "transfer-amatica",
	Run: func(cmd *cobra.Command, args []string) {
		if err := xferAmatica(); err != nil {
			panic(err)
		}
	},
}

func xferAmatica() error {
	return nil
}
