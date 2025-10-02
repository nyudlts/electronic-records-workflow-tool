package cmd

import (
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//go:embed adoc-config.yml
var vfs embed.FS

var rootCmd = &cobra.Command{}

const VERSION = "v1.1.0b.0"

// common flags
var (
	aipLoc           string
	aipFileLoc       string
	aspaceEnv        string
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
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
