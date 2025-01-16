package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rstarXfrCmd.Flags().StringVar(&ersLoc, "aips-location", "ers/", "location of AIPS to transfer to r*")
	rstarXfrCmd.Flags().StringVar(&collectionCode, "collection-code", ".*", "")
	rootCmd.AddCommand(rstarXfrCmd)
}

var rstarXfrCmd = &cobra.Command{
	Use:   "transfer-rs",
	Short: "Transfer processed AIPS to R*",
	Run: func(cmd *cobra.Command, args []string) {
		if err := transferToRstar(); err != nil {
			panic(err)
		}
	},
}

func transferToRstar() error {
	ers, err := os.Stat(ersLoc)
	if err != nil {
		return err
	}

	if !ers.IsDir() {
		return fmt.Errorf("%s is not a location", ersLoc)
	}

	directoryEntries, err := os.ReadDir(ersLoc)
	if err != nil {
		return err
	}

	xferLog := fmt.Sprintf("%S-adoc-transfer-rs.txt", collectionCode)
	_, err = os.Create(xferLog)
	if err != nil {
		return err
	}

	for _, entry := range directoryEntries {
		fmt.Println("transferring", entry.Name())
		xferBag := filepath.Join(ersLoc, entry.Name())
		xferCmd := exec.Command("rstar-scp.exp", xferBag)
		cmdOutput, err := xferCmd.CombinedOutput()
		if err != nil {
			return err
		}
		cmdOutput = append(cmdOutput, []byte("\n")...)

		f, err := os.OpenFile(xferLog, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		if _, err = f.Write(cmdOutput); err != nil {
			panic(err)
		}
	}
	return nil
}
