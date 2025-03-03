package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	bagit "github.com/nyudlts/go-bagit"
	"github.com/spf13/cobra"
)

var full bool

func init() {
	validateERsCmd.Flags().StringVar(&ersLoc, "aips-location", "aips", "location of AIPS to validate")
	validateERsCmd.Flags().BoolVar(&full, "full", false, "do a full validation instead of fast validation")
	aipCmd.AddCommand(validateERsCmd)
}

var validateERsCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate AIPS prior to transfer to R*",
	Run: func(cmd *cobra.Command, args []string) {

		//load project config
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		fmt.Printf("ADOC AIP VALIDATE %s\n", version)

		//validate the AIPS
		if err := validateAIPs(); err != nil {
			panic(err)
		}
		fmt.Println("All AIPs are valid")
	},
}

func validateAIPs() error {
	ers, err := os.Stat(ersLoc)
	if err != nil {
		return err
	}

	if !ers.IsDir() {
		return fmt.Errorf("%s is not a location", ersLoc)
	}

	logFile, err := os.Create(fmt.Sprintf("logs/%s-aip-validation.log", adocConfig.CollectionCode))
	if err != nil {
		return err
	}

	defer logFile.Close()
	log.SetOutput(logFile)

	directoryEntries, err := os.ReadDir(ersLoc)
	if err != nil {
		return err
	}

	for _, entry := range directoryEntries {
		if entry.IsDir() {

			erPath := filepath.Join(ersLoc, entry.Name())
			bag, err := bagit.GetExistingBag(erPath)
			if err != nil {
				return err
			}

			if full {
				fmt.Printf("  * validating %s\n", entry.Name())
				if err := bag.ValidateBag(false, false); err != nil {
					return err
				}
			} else {
				fmt.Printf("  * fast validating %s\n", entry.Name())
				if err := bag.ValidateBag(true, false); err != nil {
					return err
				}
			}

		}
	}

	return nil
}
