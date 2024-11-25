package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

func init() {
	rstarXfrCmd.Flags().StringVar(&ersLoc, "staging-location", ".", "location of AIPS to transfer to r*")
	rstarXfrCmd.Flags().StringVar(&ersRegex, "regexp", ".*", "regexp of files to match")
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

	if ersRegex == "" {
		return fmt.Errorf("regexp cannot not be nil")
	}

	ersPtn := regexp.MustCompile(fmt.Sprintf("%s", ersRegex))

	for _, entry := range directoryEntries {
		if entry.IsDir() && ersPtn.MatchString(entry.Name()) {
			fmt.Println("transferring", entry.Name())
			xferBag := filepath.Join(ersLoc, entry.Name())
			xferCmd := exec.Command("rstar-scp.exp", xferBag)
			xferCmd.Stdout = os.Stdout
			if err := xferCmd.Run(); err != nil {
				return err
			}
		}
	}
	return nil
}
