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
	clamCmd.Flags().StringVar(&ersLoc, "staging-location", "", "location of directories to run clamav on")
	clamCmd.Flags().StringVar(&ersRegex, "regexp", ".*", "regular expression of files to run clamav on")
	rootCmd.AddCommand(clamCmd)
}

var clamCmd = &cobra.Command{
	Use:   "clamscan",
	Short: "Run clamav against a package",
	Run: func(cmd *cobra.Command, args []string) {
		if err := clamscan(); err != nil {
			panic(err)
		}
	},
}

func clamscan() error {
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
			fmt.Printf("Scanning %s for viruses\n", entry.Name())
			xfer := filepath.Join(ersLoc, entry.Name())
			logName := filepath.Join(ersLoc, "metadata", fmt.Sprintf("%s_clamscan.log", entry.Name()))
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
