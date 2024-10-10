package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	bagit "github.com/nyudlts/go-bagit"
	"github.com/spf13/cobra"
)

func init() {
	validateERsCmd.Flags().StringVar(&ersLoc, "staging-location", "", "")
	validateERsCmd.Flags().StringVar(&ersRegex, "regexp", "", "")
	rootCmd.AddCommand(validateERsCmd)
}

var validateERsCmd = &cobra.Command{
	Use: "validate-ers",
	Run: func(cmd *cobra.Command, args []string) {
		if err := validateERs(); err != nil {
			panic(err)
		}
		fmt.Println("All ER bags are valid")
	},
}

func validateERs() error {
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
			fmt.Printf("fast validating %s\n", entry.Name())
			erPath := filepath.Join(ersLoc, entry.Name())
			bag, err := bagit.GetExistingBag(erPath)
			if err != nil {
				return err
			}

			if err := bag.ValidateBag(true, false); err != nil {
				log.Printf("[ERROR] %s not valid according to Payload-Oxum", erPath)
			} else {
				log.Printf("[INFO] %s valid according to Payload-Oxum", erPath)
			}
		}
	}

	return nil
}
