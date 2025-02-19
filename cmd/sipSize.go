package cmd

import (
	"github.com/spf13/cobra"
)

var directoryStats bool

func init() {
	sipSizeCmd.Flags().BoolVarP(&directoryStats, "directory", "d", false, "Print size info for each directory")
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

		//print the total size of SIP
		if err := getPackageSize(adocConfig.StagingLoc); err != nil {
			panic(err)
		}

		//print the stats of each directory if flag set
		if directoryStats {
			if err := printDirectoryStats(adocConfig.StagingLoc); err != nil {
				panic(err)
			}
		}
	},
}
