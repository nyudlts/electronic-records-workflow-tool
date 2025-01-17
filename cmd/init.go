package cmd

import (
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
	CollectionCall string `yaml:"collection-call"`
	ProjectLoc     string `yaml:"project-location"`
}

func init() {
	initCmd.Flags().StringVar(&partnerCode, "partner-code", "", "the partner code to use for the adoc")
	initCmd.Flags().StringVar(&collectionCall, "collection-call", "", "the collection call no. to use for the adoc")
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use: "init",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initAdoc(); err != nil {
			panic(err)
		}
	},
}

func initAdoc() error {
	var err error
	adocConfig, err = loadConfig()
	if err != nil {
		return err
	}

	if err := mkProjectDir(); err != nil {
		return err
	}
	return nil
}

func loadConfig() (*AdocConfig, error) {
	b, err := vfs.ReadFile("adoc-config.yml")
	if err != nil {
		return nil, err
	}
	config := &AdocConfig{}
	if err := yaml.Unmarshal(b, config); err != nil {
		return nil, err
	}

	config.PartnerCode = partnerCode
	config.StagingLoc = filepath.Join(config.StagingLoc, partnerCode, collectionCall)
	config.CollectionCall = collectionCall
	config.ProjectLoc = filepath.Join(config.ProjectLoc, config.CollectionCall)
	return config, nil
}

func mkProjectDir() error {
	if err := os.Mkdir(adocConfig.ProjectLoc, 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, adocConfig.PartnerCode, "ers"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, adocConfig.PartnerCode, "logs"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, adocConfig.PartnerCode, "rsync"), 0775); err != nil {
		return err
	}

	b, err := yaml.Marshal(adocConfig)
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(adocConfig.ProjectLoc, "config.yml"), b, 0755); err != nil {
		return err
	}

	return nil
}
