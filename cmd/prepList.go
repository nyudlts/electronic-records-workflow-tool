package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	cp "github.com/otiai10/copy"
	"github.com/spf13/cobra"
)

func init() {
	listCmd.Flags().StringVar(&aipFileLoc, "aip-file", "", "")
	listCmd.Flags().StringVar(&stagingLoc, "staging-location", "", "")
	listCmd.Flags().StringVar(&tmpLoc, "tmp-location", "", "")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use: "prep-list",
	Run: func(cmd *cobra.Command, args []string) {
		processList()
	},
}

func processList() error {
	fmt.Println(aipFileLoc)
	aipFile, err := os.Open(aipFileLoc)
	if err != nil {
		return err
	}
	defer aipFile.Close()
	scanner := bufio.NewScanner(aipFile)

	for scanner.Scan() {
		aipLocation := scanner.Text()
		fmt.Println(aipLocation)
		fi, err := os.Stat(aipLocation)
		if err != nil {
			return err
		}

		//set copy options
		options.PreserveTimes = true
		options.PermissionControl = cp.AddPermission(0755)

		//copy the directory to the staging area
		aipLoc = filepath.Join(stagingLoc, fi.Name())
		fmt.Printf("\nCopying package from %s to %s\n", aipLocation, aipLoc)
		if err := cp.Copy(aipLoc, aipLoc, options); err != nil {
			return err
		}

		//run the update process
		if tmpLoc == "" {
			tmpLoc = "/tmp"
		}

		fmt.Printf("\nUpdating package at %s\n", aipLoc)
		if err := prepPackage(aipLoc, tmpLoc); err != nil {
			return err
		}

	}

	return nil
}
