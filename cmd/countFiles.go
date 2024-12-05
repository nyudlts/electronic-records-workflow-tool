package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	countCmd.Flags().StringVar(&stagingLoc, "staging-location", ".", "the location of directories to count files in")
	rootCmd.AddCommand(countCmd)
}

var wrtr bufio.Writer

var countCmd = &cobra.Command{
	Use: "count",
	Run: func(cmd *cobra.Command, args []string) {
		if err := countFiles(); err != nil {
			panic(err)
		}
	},
}

func countFiles() error {
	oFile, _ := os.Create("adoc-counts.tsv")
	defer oFile.Close()
	wrtr = *bufio.NewWriter(oFile)
	wrtr.WriteString(fmt.Sprintf("%s\t%d\n", "path", "count"))
	objs, err := os.ReadDir(stagingLoc)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		if obj.IsDir() {
			if err := getFileCount(filepath.Join(stagingLoc, obj.Name())); err != nil {
				return err
			}
		}
	}
	wrtr.Flush()
	return nil
}

func getFileCount(path string) error {
	count := 0
	err := filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		return err
	}
	wrtr.WriteString(fmt.Sprintf("%s\t%d\n", filepath.Base(path), count))
	return nil
}
