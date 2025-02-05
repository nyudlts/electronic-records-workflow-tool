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
	rootCmd.AddCommand(sourceXferCmd)
}

var sourceXferCmd = &cobra.Command{
	Use: "transfer-source",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("adoc %s transfer-acm\n", version)
		//load the configuration
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		//run
		if err := transferACM(); err != nil {
			panic(err)
		}
	},
}

func transferACM() error {

	//create the logfile
	logFileName := filepath.Join(adocConfig.LogLoc, "rsync", fmt.Sprintf("%s-adoc-acm-transfer-rsync.txt", adocConfig.CollectionCode))
	logFile, err := os.Create(logFileName)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(logFile)
	defer logFile.Close()

	cmd := exec.Command("rsync", "-rav", fmt.Sprintf("%s", adocConfig.SourceLoc), adocConfig.StagingLoc)
	fmt.Printf("copying %s to %s\n", adocConfig.SourceLoc, adocConfig.StagingLoc)

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
