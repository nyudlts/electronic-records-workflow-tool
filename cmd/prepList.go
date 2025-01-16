package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	listCmd.Flags().StringVar(&aipFileLoc, "aip-file", "aip-file.txt", "the location of the aip-file containing aips to process")
	listCmd.Flags().StringVar(&stagingLoc, "aip-location", "ers/", "location to stage aips")
	listCmd.Flags().StringVar(&tmpLoc, "tmp-location", ".", "location to store tmp bag-info.txt")
	listCmd.Flags().StringVar(&collectionCode, "collection-code", "", "the collection code for the aips")
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "prep-aips",
	Short: "Prepare a list of AIPs for transfer to R*",
	Run: func(cmd *cobra.Command, args []string) {
		processList()
	},
}

func processList() error {
	aipFileLoc := fmt.Sprintf("%s-%s", collectionCode, aipFileLoc)
	aipFile, err := os.Open(aipFileLoc)
	if err != nil {
		return err
	}
	defer aipFile.Close()
	scanner := bufio.NewScanner(aipFile)

	logFile, err := os.Create(fmt.Sprintf("%s-adoc-prep-aips.log", collectionCode))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	for scanner.Scan() {
		aipLocation := scanner.Text()
		fi, err := os.Stat(aipLocation)
		if err != nil {
			return err
		}

		msg := fmt.Sprintf("updating %s\n", fi.Name())
		fmt.Println(msg)
		log.Println("INFO", msg)

		//copy the directory to the staging area
		aipStageLoc := filepath.Join(stagingLoc, fi.Name())
		msg = fmt.Sprintf("Copying package from %s to %s", aipLocation, aipLoc)
		fmt.Println(msg)
		log.Printf("[INFO] %s", msg)
		cmd := exec.Command("rsync", "-rav", aipLocation, stagingLoc)

		b, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}

		if err := os.WriteFile(fmt.Sprintf("%s-rsync-output.txt", fi.Name()), b, 0666); err != nil {
			return err
		}

		fmt.Println("OK")

		msg = fmt.Sprintf("Updating package at %s", aipLoc)
		fmt.Print(msg + ": ")
		log.Printf("[INFO] %s", msg)
		if err := prepPackage(aipStageLoc, tmpLoc); err != nil {
			return err
		}
		fmt.Println("OK")

	}

	return nil
}
