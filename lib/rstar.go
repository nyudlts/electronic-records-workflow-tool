package lib

import "fmt"

func PrintRStarPackageSize(directories bool) error {
	fmt.Println("ewt rstar size, version", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	if err := getPackageSize(config.AIPLoc); err != nil {
		return err
	}

	if directories {
		if err := printDirectoryStats(config.AIPLoc); err != nil {
			return err
		}
	}

	return nil
}

func PrepareRStarPackages() error {
	fmt.Println("ewt rstar prep packages", VERSION)
	return nil
}

func PrepareSinglePackage(path string) error {
	fmt.Println("ewt rstar prep single package", VERSION)
	return nil
}

func ValidateRStarPackages() error {
	fmt.Println("ewt rstar validate", VERSION)
	return nil
}

func TransferRStarPackages() error {
	fmt.Println("ewt rstar transfer", VERSION)
	return nil
}
