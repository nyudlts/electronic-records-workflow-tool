package cmd

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

//go:embed adoc-config.yml
var vfs embed.FS

var rootCmd = &cobra.Command{}

const version = "v1.0.0"

// common flags
var (
	aipLoc           string
	aipFileLoc       string
	sourceLoc        string
	stagingLoc       string
	tmpLoc           string
	amaticaConfigLoc string
	ersLoc           string
	pollTime         int
	collectionCode   string
	adocConfig       *AdocConfig
	projectLoc       string
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadProjectConfig() error {
	//read the adoc-config
	b, err := os.ReadFile("config.yml")
	if err != nil {
		return err
	}

	//unmarshal to config options
	if err := yaml.Unmarshal(b, &adocConfig); err != nil {
		return err
	}

	return nil
}

func printConfig() error {
	b, err := json.MarshalIndent(adocConfig, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}
