package lib

import "fmt"

func PrintXferPackageSize(directories bool) error {
	fmt.Println("ewt amatica size, version", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	if err := getPackageSize(config.XferLoc); err != nil {
		return err
	}

	if directories {
		if err := printDirectoryStats(config.XferLoc); err != nil {
			return err
		}
	}

	return nil
}
