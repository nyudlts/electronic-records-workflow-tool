package cmd

import "github.com/spf13/cobra"

func init() {
	amaticaSizeCmd.Flags().BoolVarP(&directoryStats, "directory", "d", false, "Print size info for each directory")
	amaticaCmd.AddCommand(aipSizeCmd)
}

var amaticaSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Get the file count and size of an XFER package",
	Run: func(cmd *cobra.Command, args []string) {

		//load the project config
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		//print the total size of SIP
		if err := getPackageSize(adocConfig.XferLoc); err != nil {
			panic(err)
		}

		//print the stats of each directory if flag set
		if directoryStats {
			if err := printDirectoryStats(adocConfig.XferLoc); err != nil {
				panic(err)
			}
		}
	},
}
