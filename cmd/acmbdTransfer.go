package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var sourceLocation string
var stagingLocation string
var collectionCode string

func init() {
	acmbdXferCmd.Flags().StringVar(&sourceLocation, "source-location", "", "Source location")
	acmbdXferCmd.Flags().StringVar(&stagingLocation, "staging-location", "", "Staging location")
	acmbdXferCmd.Flags().StringVar(&collectionCode, "collection-code", "", "Collection code")
	rootCmd.AddCommand(acmbdXferCmd)
}

var acmbdXferCmd = &cobra.Command{
	Use: "transfer-acm",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("adoc %s transfer-acm\n", version)
		if err := transferACM(); err != nil {
			panic(err)
		}
	},
}

func transferACM() error {

	//create the logfile
	logFileName := fmt.Sprintf("%s-adoc-acm-transfer.txt", collectionCode)
	logFile, err := os.Create(logFileName)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(logFile)
	defer logFile.Close()

	targetDir := filepath.Join(stagingLocation, collectionCode)
	if err := os.Mkdir(targetDir, 0775); err != nil {
		return err
	}

	sourceLocationFiles := fmt.Sprintf("%s/*", sourceLocation)

	cmd := exec.Command("rsync", "-rav", sourceLocationFiles, targetDir)
	fmt.Printf("copying %s to %s\n", sourceLocation, targetDir)

	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	if _, err := writer.Write(b); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}
