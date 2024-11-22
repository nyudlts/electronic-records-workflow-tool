package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	listCmd.Flags().StringVar(&aipFileLoc, "aip-file", "", "")
	listCmd.Flags().StringVar(&stagingLoc, "staging-location", "", "")
	listCmd.Flags().StringVar(&tmpLoc, "tmp-location", "", "")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "prep-list",
	Short: "Prepare a list of AIPs for transfer to R*",
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

		fmt.Println(fi.Name())

		//copy the directory to the staging area
		aipStageLoc := filepath.Join(stagingLoc, fi.Name())
		fmt.Printf("\nCopying package from %s to %s\n", aipLocation, aipLoc)
		cmd := exec.Command("rsync", "-rav", aipLocation, stagingLoc)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}

		fmt.Printf("\nUpdating package at %s\n", aipLoc)
		if err := prepPackage(aipStageLoc, tmpLoc); err != nil {
			return err
		}

	}

	return nil
}
