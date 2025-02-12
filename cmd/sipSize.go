package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	sipCmd.AddCommand(sipSizeCmd)
}

var sipSizeCmd = &cobra.Command{
	Use:   "size",
	Short: "Get the size  and number of files in a SIP",
	Run: func(cmd *cobra.Command, args []string) {
		if err != loadProjectConfig(); err != nil {
			panic(err)
		}

		if err := getSipSize(); err != nil {
			panic(err)
		}
	},
}

func getSipSize() error {
	numFiles := 0
	numDirectories := 0
	sizeFiles := int64(0)

	if err := filepath.Walk(adocConfig.StagingLoc, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			numDirectories++
		} else {
			numFiles++
			sizeFiles += info.Size()
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}
