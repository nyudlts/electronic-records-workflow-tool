package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

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

	if err := loadConfig(); err != nil {
		return err
	}

	//create a log file
	logFile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s-rstar-prep.log", config.CollectionCode)))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//get the aip file
	aipFileLoc := fmt.Sprintf("%s-aip-file.txt", config.CollectionCode)
	if _, err := os.Stat(aipFileLoc); err != nil {
		log.Printf("[ERROR] aip file %s not found: %v", aipFileLoc, err)
		return fmt.Errorf("aip file %s not found: %v", aipFileLoc, err)
	}

	//open the aipFile
	aipFile, err := os.Open(aipFileLoc)
	if err != nil {
		return err
	}
	defer aipFile.Close()
	scanner := bufio.NewScanner(aipFile)

	for scanner.Scan() {
		aipLocation := scanner.Text()
		fi, err := os.Stat(aipLocation)
		if err != nil {
			log.Printf("[ERROR] aip package %s does not exist: %v", aipFileLoc, err)
			return fmt.Errorf("aip package %s does not exist: %v", aipFileLoc, err)
		}

		msg := fmt.Sprintf("updating %s", fi.Name())
		fmt.Println(msg)
		log.Println("[INFO]", msg)
	}

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
