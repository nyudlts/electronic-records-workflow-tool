package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
		if err := getSipSize(); err != nil {
			panic(err)
		}

		//print the stats of each directory if flag set
		if directoryStats {
			if err := printDirectoryStats(); err != nil {
				panic(err)
			}
		}
	},
}

type DirectoryStats struct {
	Size           int64
	NumFiles       int
	NumDirectories int
}

func printDirectoryStats() error {
	directoryStats := []DirectoryStats{}
	dirs, err := os.ReadDir(adocConfig.StagingLoc)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		ds := DirectoryStats{0, 0, 0}
		if err := filepath.Walk(filepath.Join(adocConfig.StagingLoc, dir.Name()), func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				ds.NumDirectories++
			} else {
				ds.NumFiles++
				ds.Size += info.Size()
			}
			return nil
		}); err != nil {
			return err
		}
	}

	for _, ds := range directoryStats {
		fmt.Printf("%s: %d files in %d directories (%d bytes)\n", path, ds.NumFiles, ds.NumDirectories, ds.Size)
	}

	return nil

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

	fmt.Printf("%d files in %d directories (%d bytes)\n", numFiles, numDirectories, sizeFiles)
	return nil
}
