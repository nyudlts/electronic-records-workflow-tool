package cmd

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	listCmd.Flags().StringVar(&aipFileLoc, "aip-file", "", "the location of the aip-file containing aips to process")
	listCmd.Flags().StringVar(&stagingLoc, "aip-location", "aips/", "location to stage aips")
	listCmd.Flags().StringVar(&tmpLoc, "tmp-location", "logs", "location to store tmp bag-info.txt")
	aipCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "prep",
	Short: "Prepare a list of AIPs for transfer to R*",
	Run: func(cmd *cobra.Command, args []string) {

		//load the project configuration
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		fmt.Printf("ADOC AIP PREP %s\n", version)

		//locate the aip file
		if err := locateAIPFile(); err != nil {
			panic(err)
		}

		if err := processList(); err != nil {
			panic(err)
		}
	},
}

func locateAIPFile() error {
	if aipFileLoc != "" {
		return nil
	}

	logFiles, err := os.ReadDir(adocConfig.LogLoc)
	if err != nil {
		return err
	}

	for _, logFile := range logFiles {
		if strings.Contains(logFile.Name(), "aip-file.txt") {
			aipFileLoc = filepath.Join(adocConfig.LogLoc, logFile.Name())
			return nil
		}
	}

	return fmt.Errorf("aip-file.txt not found in %s", adocConfig.LogLoc)
}

func processList() error {
	if aipFileLoc == "" {
		aipFileLoc = fmt.Sprintf("%s-aip-file.txt", adocConfig.CollectionCode)
	}
	aipFile, err := os.Open(aipFileLoc)
	if err != nil {
		return err
	}
	defer aipFile.Close()
	scanner := bufio.NewScanner(aipFile)

	logFile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s-adoc-aip.log", adocConfig.CollectionCode)))
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

		msg := fmt.Sprintf("updating %s", fi.Name())
		fmt.Println(msg)
		log.Println("INFO", msg)

		//copy the directory to the staging area
		aipStageLoc := filepath.Join(stagingLoc, fi.Name())
		msg = fmt.Sprintf("Copying package from %s to %s", aipLocation, "aips")
		fmt.Println(msg)
		log.Printf("[INFO] %s", msg)
		cmd := exec.Command("rsync", "-rav", aipLocation, stagingLoc)

		b, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}

		if err := os.WriteFile(filepath.Join("logs", "rsync", fmt.Sprintf("%s-rsync-output.txt", fi.Name())), b, 0775); err != nil {
			return err
		}

		fmt.Println("OK")

		msg = fmt.Sprintf("Updating package at %s\n", aipLocation)
		fmt.Print(msg + ": ")
		log.Printf("[INFO] %s", msg)
		if err := prepPackage(aipStageLoc, tmpLoc); err != nil {
			return err
		}
		fmt.Println("OK")

	}

	return nil
}
