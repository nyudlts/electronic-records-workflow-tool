package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	sipCmd.AddCommand(clamCmd)
}

var clamCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run clamav against a package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		fmt.Println("ADOC SIP Scan")
		if err := clamscan(); err != nil {
			panic(err)
		}
	},
}

func clamscan() error {
	ers, err := os.Stat(adocConfig.StagingLoc)
	if err != nil {
		return err
	}

	if !ers.IsDir() {
		return fmt.Errorf("%s is not a location", adocConfig.StagingLoc)
	}

	directoryEntries, err := os.ReadDir(adocConfig.StagingLoc)
	if err != nil {
		return err
	}

	for _, entry := range directoryEntries {
		if entry.IsDir() && entry.Name() != "metadata" {
			fmt.Printf("  * Scanning %s for viruses\n", entry.Name())
			xfer := filepath.Join(adocConfig.StagingLoc, entry.Name())
			logName := filepath.Join(adocConfig.StagingLoc, "metadata", fmt.Sprintf("%s_clamscan.log", entry.Name()))
			if _, err := os.Create(logName); err != nil {
				return err
			}
			clamscanCmd := exec.Command("clamscan", "-r", xfer)
			cmdOut, err := clamscanCmd.CombinedOutput()
			if err != nil {
				return err
			}

			if err := os.WriteFile(logName, cmdOut, 0644); err != nil {
				return err
			}

		}
	}
	return nil
}
