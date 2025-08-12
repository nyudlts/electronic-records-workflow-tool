package lib

import (
	"os"

	"gopkg.in/yaml.v2"
)

var config = Config{}

func loadConfig() error {
	//read the adoc-config
	b, err := os.ReadFile("config.yml")
	if err != nil {
		return err
	}

	//unmarshal to config options
	if err := yaml.Unmarshal(b, &config); err != nil {
		return err
	}

	return nil
}
