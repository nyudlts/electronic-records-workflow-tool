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
	pollTime         int
	windows          bool
	collectionCode   string
	xferLoc          string
	partnerCode      string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
