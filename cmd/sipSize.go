package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	sipSizeCmd.Flags().BoolVarP(&directories, "directory", "d", false, "Print size info for each directory")
	sipCmd.AddCommand(sipSizeCmd)
}

var sipSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Get the size  and number of files in a SIP",
	Run: func(cmd *cobra.Command, args []string) {
		//load the project config
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		fmt.Printf("ADOC SIP size %s\n", version)

		//print the total size of SIP
		if err := getPackageSize(adocConfig.SIPLoc); err != nil {
			panic(err)
		}

		//print the stats of each directory if flag set
		if directories {
			if err := printDirectoryStats(adocConfig.SIPLoc); err != nil {
				panic(err)
			}
		}
	},
}
