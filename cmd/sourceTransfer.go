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
	sourceCmd.AddCommand(sourceXferCmd)
}

var sourceXferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Transfer source files to the SIP directory",
	Run: func(cmd *cobra.Command, args []string) {

		//load the configuration
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		fmt.Printf("ADOC source transfer %s\n", version)

		//run
		if err := transferSIP(); err != nil {
			panic(err)
		}

		//check if there is a metadata directory in SIP, if not create it
		if err := checkMDDir(); err != nil {
			panic(err)
		}

	},
}

func transferSIP() error {

	//create the logfile
	logFileName := filepath.Join(adocConfig.LogLoc, "rsync", fmt.Sprintf("%s-source-transfer-rsync.txt", adocConfig.CollectionCode))
	logFile, err := os.Create(logFileName)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(logFile)
	defer logFile.Close()

	fmt.Printf("  * Transferring %s to sip directory\n", adocConfig.SourceLoc)
	cmd := exec.Command("rsync", "-rav", adocConfig.SourceLoc, "sip")

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

	fmt.Printf("  * Transfer complete\n")

	return nil
}

func checkMDDir() error {
	mdDir := filepath.Join(adocConfig.SIPLoc, "metadata")
	if _, err := os.Stat(mdDir); os.IsNotExist(err) {
		fmt.Printf("  * Creating metadata directory in SIP\n")
		if err := os.Mkdir(mdDir, 0755); err != nil {
			return err
		}
	}
	return nil
}
