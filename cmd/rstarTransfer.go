package cmd

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	validateERsCmd.Flags().StringVar(&ersLoc, "staging-location", "", "")
	validateERsCmd.Flags().StringVar(&ersRegex, "regexp", "", "")
	rootCmd.AddCommand(rstarXfrCmd)
}
var rstarXfrCmd = &cobra.Command{
	Use: "transfer",
	Run: func(cmd *cobra.Command, args []string) {
		if err := transferToRstar(); err != nil {
			panic(err)
		}
	},
}

func transferToRstar() err {
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

	ersPtn := regexp.MustCompile(fmt.Sprintf(".*%s*", ersRegex))

	for _, entry := range directoryEntries {
		if entry.IsDir() && ersPtn.MatchString(entry.Name()) {
			fmt.Println("transferring", entry.Name())
			xferBag := filepath.Join(ersLoc,entry.Name)
			xferCmd := exec.Command("rstar-scp.exp", xferBag)
			xferCmd.Stdout = os.Stdout
			if err := xferCmd.Run(); err != nil {
				return err
			}
		}

	return nil
}
