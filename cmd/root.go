package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

const version = "v0.2.0"

var (
	aipLoc          string
	aipFileLoc      string
	sourceLoc       string
	stagingLoc      string
	tmpLoc          string
	aspaceConfigLoc string
	aspaceWOLoc     string
	aspaceEnv       string
	ersLoc          string
	ersRegex        string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
