package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type AdocConfig struct {
	StagingLoc       string `yaml:"staging-location"`
	SourceLoc        string `yaml:"source-location"`
	XferLoc          string `yaml:"xfer-location"`
	PartnerCode      string `yaml:"partner-code"`
	CollectionCode   string `yaml:"collection-code"`
	ProjectLoc       string `yaml:"project-location"`
	LogLoc           string `yaml:"log-location"`
	AIPLoc           string `yaml:"aip-location"`
	AMTransferSource string `yaml:"archivematica-transfer-source"`
}

func init() {
	initCmd.Flags().StringVarP(&partnerCode, "partner-code", "p", "", "the partner code to use adoc")
	initCmd.Flags().StringVarP(&collectionCode, "collection-code", "c", "", "the collection code to use for adoc")
	initCmd.Flags().StringVarP(&sourceLoc, "source-location", "s", "", "the source location for the collection")
	initCmd.Flags().StringVarP(&projectLoc, "project-location", "l", "", "the project location for the collection")
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
	//load the configuration
	adocConfig, err = loadConfig()
	if err != nil {
		return err
	}

	if err := printConfig(); err != nil {
		return err
	}

	//create default project directories
	if err := mkProjectDir(); err != nil {
		return err
	}

	//write the adoc-config to the project directory
	if err := writeAdocConfig(); err != nil {
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
	config.StagingLoc = "sip"
	config.AIPLoc = "aips"
	config.LogLoc = "logs"
	config.CollectionCode = collectionCode
	if projectLoc != "" {
		config.ProjectLoc = filepath.Join(projectLoc, collectionCode)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		config.ProjectLoc = filepath.Join(wd, collectionCode)
	}
	config.SourceLoc = sourceLoc

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
	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, "logs", "rsync"), 0775); err != nil {
		return err
	}

	//create the sip output directory
	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, "sip"), 0775); err != nil {
		return err
	}

	//create the xfer directory
	if err := os.Mkdir(filepath.Join(adocConfig.ProjectLoc, "xfer"), 0775); err != nil {
		return err
	}

	return nil
}

func writeAdocConfig() error {

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
