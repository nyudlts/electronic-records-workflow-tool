package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{}

const version = "v0.2.0"

var (
	sourceLoc  string
	stagingLoc string
	aipLoc     string
	tmpLoc     string
	aipFileLoc string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
