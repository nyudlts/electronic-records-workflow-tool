package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

const version = "v0.2.2"

// common flags
var (
	aipLoc           string
	aipFileLoc       string
	sourceLoc        string
	stagingLoc       string
	tmpLoc           string
	amaticaConfigLoc string
	aspaceConfigLoc  string
	aspaceWOLoc      string
	aspaceEnv        string
	ersLoc           string
	ersRegex         string
	pollTime         int
	windows          bool
	collectionCode   string
	xferLoc          string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
