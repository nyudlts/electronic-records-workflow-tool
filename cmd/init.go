package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var adocConfig *AdocConfig

type AdocConfig struct {
	StagingLoc     string `yaml:"staging-location"`
	XferLoc        string `yaml:"xfer-location"`
	PartnerCode    string `yaml:"partner-code"`
	CollectionCode string `yaml:"collection-code"`
	ProjectLoc     string `yaml:"project-location"`
}

func init() {
	initCmd.Flags().StringVarP(&partnerCode, "partner-code", "p", "", "the partner code to use adoc")
	initCmd.Flags().StringVarP(&collectionCode, "collection-code", "c", "", "the collection code to use for adoc")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use: "init",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("partner-code: %s\n", partnerCode)
		fmt.Printf("collection-code: %s\n", collectionCode)

		if err := initAdoc(); err != nil {
			panic(err)
		}
	},
}

func initAdoc() error {
	var err error
	//load the configuration
	adocConfig, err = loadConfig()
	if err != nil {
		return err
	}

	//create default project directories
	if err := mkProjectDir(); err != nil {
		return err
	}
	return nil
}

func loadConfig() (*AdocConfig, error) {
	//read the adoc-config
	b, err := vfs.ReadFile("adoc-config.yml")
	if err != nil {
		return nil, err
	}

	//unmarshal to config options
	config := &AdocConfig{}
	if err := yaml.Unmarshal(b, config); err != nil {
		return nil, err
	}

	//update members
	config.PartnerCode = partnerCode
	config.StagingLoc = filepath.Join(config.StagingLoc, partnerCode, collectionCode)
	config.CollectionCode = collectionCode
	config.ProjectLoc = filepath.Join(config.ProjectLoc, config.CollectionCode)

	return config, nil
}

func mkProjectDir() error {
	//create the project directory
	if err := os.Mkdir(adocConfig.ProjectLoc, 0775); err != nil {
		return err
	}

	//create the aips directory
	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, "aips"), 0775); err != nil {
		return err
	}

	//create the logs directory
	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, "logs"), 0775); err != nil {
		return err
	}

	//create the resync output directory
	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, "rsync"), 0775); err != nil {
		return err
	}

	//marshall the updated config
	b, err := yaml.Marshal(adocConfig)
	if err != nil {
		return err
	}

	//write the config to the project directory
	if err := os.WriteFile(filepath.Join(adocConfig.ProjectLoc, "config.yml"), b, 0755); err != nil {
		return err
	}

	return nil
}
