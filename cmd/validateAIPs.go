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

var full bool

func init() {
	validateERsCmd.Flags().StringVar(&ersLoc, "staging-location", ".", "location of AIPS to validate")
	validateERsCmd.Flags().StringVar(&ersRegex, "regexp", "", "regexp for directories tp validate.")
	validateERsCmd.Flags().BoolVar(&full, "full", false, "do a full validation instead of fast validation")
	validateERsCmd.Flags().StringVar(&collectionCode, "collection-code", "", "collection code to validate")
	rootCmd.AddCommand(validateERsCmd)
}

var validateERsCmd = &cobra.Command{
	Use:   "validate-aips",
	Short: "Validate AIPS prior to transfer to R*",
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

	logFile, err := os.Create(fmt.Sprintf("%s-adoc-aip-validation.log", collectionCode))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	directoryEntries, err := os.ReadDir(ersLoc)
	if err != nil {
		return err
	}

	var ersPtn *regexp.Regexp
	if ersRegex != "" {
		ersPtn = regexp.MustCompile(fmt.Sprintf(".*%s*", ersRegex))
	}

	for _, entry := range directoryEntries {
		if entry.IsDir() {
			if ersPtn == nil || ersPtn.MatchString(entry.Name()) {
				erPath := filepath.Join(ersLoc, entry.Name())
				bag, err := bagit.GetExistingBag(erPath)
				if err != nil {
					return err
				}

				if full {
					fmt.Printf("validating %s\n", entry.Name())
					if err := bag.ValidateBag(false, false); err != nil {
						return err
					}
				} else {
					fmt.Printf("fast validating %s\n", entry.Name())
					if err := bag.ValidateBag(true, false); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}
