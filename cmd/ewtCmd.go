package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

// common flags
var (
	aipLoc           string
	aipFileLoc       string
	aspaceEnv        string
	aspaceConfig     string
	directories      bool
	ingests          bool
	transfers        bool
	sourceLoc        string
	stagingLoc       string
	tmpLoc           string
	amaticaConfigLoc string
	ersLoc           string
	pollTime         int
	collectionCode   string
	projectLoc       string
	profile          string
	numWorkers       int
	fullValidation   bool
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("  * error encountered: %v\n", err)
		os.Exit(1)
	}
}
