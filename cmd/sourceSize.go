package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	sourceSizeCmd.Flags().BoolVarP(&directories, "directory", "d", false, "Print size info for each directory")
	sourceCmd.AddCommand(sourceSizeCmd)
}

var sourceSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Get size of source directory",
	Run: func(cmd *cobra.Command, args []string) {
		//load the project config
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		fmt.Printf("ADOC Source Size %s\n", version)

		//print the total size of source directory
		if err := getPackageSize(adocConfig.SourceLoc); err != nil {
			panic(err)
		}

		//print the stats of each directory if flag set
		if directories {
			if err := printDirectoryStats(adocConfig.SourceLoc); err != nil {
				panic(err)
			}
		}
	},
}
