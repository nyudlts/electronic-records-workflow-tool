package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type AdocConfig struct {
	StagingLoc       string `yaml:"staging-location"`
	SourceLoc        string `yaml:"source-location"`
	PartnerCode      string `yaml:"partner-code"`
	CollectionCode   string `yaml:"collection-code"`
	ProjectLoc       string `yaml:"project-location"`
	LogLoc           string `yaml:"log-location"`
	AIPLoc           string `yaml:"aip-location"`
	AMTransferSource string `yaml:"archivematica-transfer-source"`
	XferLoc          string `yaml:"xfer-location"`
}

func init() {
	initCmd.Flags().StringVarP(&collectionCode, "collection-code", "c", "", "the collection code to use for adoc")
	initCmd.Flags().StringVarP(&sourceLoc, "source-location", "s", "", "the source location for the collection")
	projectCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a ADOC transfer",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("ADOC INIT")

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

	//print the configuration
	fmt.Println("Configuration:")
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
	config.PartnerCode = strings.Split(collectionCode, "_")[0]
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
	config.StagingLoc = filepath.Join(config.ProjectLoc, "sip")
	config.AIPLoc = filepath.Join(config.ProjectLoc, "aips")
	config.LogLoc = filepath.Join(config.ProjectLoc, "logs")
	config.XferLoc = filepath.Join(config.ProjectLoc, "xfer")
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
