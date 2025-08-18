package lib

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func PrintSIPPackageSize(directories bool) error {
	fmt.Println("ewt sip size, version", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	if err := getPackageSize(config.SIPLoc); err != nil {
		return err
	}

	if directories {
		if err := printDirectoryStats(config.SIPLoc); err != nil {
			return err
		}
	}

	return nil
}

func CleanSip() error {
	fmt.Println("ewt sip clean, version", VERSION)

	//load the project configuration
	if err := loadConfig(); err != nil {
		return err
	}
	deleteCount := 0
	if err := filepath.Walk(config.SIPLoc, func(path string, info fs.FileInfo, err error) error {

		if !info.IsDir() {
			if info.Name() == ".DS_Store" || info.Name() == "Thumbs.db" {
				if err := os.Remove(path); err != nil {
					return err
				}
				fmt.Printf("  * deleted %s\n", path)
				deleteCount++
			}
		}

		return nil
	}); err != nil {
		return err
	}
	fmt.Printf("  * %d files deleted\n", deleteCount)
	return nil
}

func GenerateTransferInfo(profile string) error {
	fmt.Println("ewt sip gen transfer, version", VERSION)
	fmt.Println("  * generating transfer info for profile:", profile)

	if err := loadConfig(); err != nil {
		return err
	}

	profileFilename := strings.ToUpper(profile) + ".txt"
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	parentDirectory, _ := filepath.Split(wd)
	profileFile := filepath.Join(parentDirectory, "templates", profileFilename)

	if _, err := os.Stat(profileFile); err != nil {
		return err
	}

	templateFile := filepath.Join(parentDirectory, "templates", "ewt.txt")
	if _, err := os.Stat(templateFile); err != nil {
		return err
	}

	xferInfo := filepath.Join(config.SIPLoc, "metadata", "transfer-info.txt")

	outFile, err := os.Create(xferInfo)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)

	profileBytes, err := os.ReadFile(profileFile)
	if err != nil {
		return err
	}

	templateBytes, err := os.ReadFile(templateFile)
	if err != nil {
		return err
	}

	writer.Write(profileBytes)
	writer.Write(templateBytes)
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func ValidateSIP() error {
	fmt.Println("ewt sip validate,", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	//create a logger
	logFile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s-sip-validate.log", config.CollectionCode)))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Printf("[INFO] ewt validate sip %s\n", VERSION)
	fmt.Printf("  * validating SIP at %s\n", config.SIPLoc)
	log.Printf("[INFO] validating SIP transfer package at %s\n", config.SIPLoc)

	//check that the source directory exists
	fmt.Print("  1. checking that SIP location exists and is a directory: ")
	fileInfo, err := os.Stat(config.SIPLoc)
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
		fmt.Printf("SIP location %s does not exist, exiting", config.SIPLoc)
		return err
	}

	if !fileInfo.IsDir() {
		log.Printf("[ERROR] %s is not a directory\n", config.SIPLoc)
		fmt.Printf("  * SIP location %s is not a directory, exiting", config.SIPLoc)
		return fmt.Errorf("%s is not a directory", config.SIPLoc)
	}
	log.Printf("[INFO] %s exists and is a directory", config.SIPLoc)
	fmt.Println(" OK")

	//check that there is a metadata directory
	fmt.Print("  2. checking that SIP directory contains a metadata directory: ")
	mdDirLocation := filepath.Join(config.SIPLoc, "metadata")
	mdDir, err := os.Stat(mdDirLocation)
	if err != nil {
		fmt.Printf("SIP location %s does not contain a metadata directory", config.SIPLoc)
		log.Printf("[ERROR] %s does not contain a metadata directory\n", config.SIPLoc)
		return (err)
	}

	if !mdDir.IsDir() {
		fmt.Printf("  * %s metadata directory is not a directory\n", mdDirLocation)
		log.Printf("[ERROR] %s is not a directory\n", mdDirLocation)
		return fmt.Errorf("[ERROR] %s is not a directory\n", mdDirLocation)

	}
	log.Printf("[INFO] %s contains a metadata directory\n", config.SIPLoc)
	fmt.Println("OK")

	//finish up
	fmt.Printf("  * Validation report written to %s\n", logFile.Name())
	return nil
}

func ScanAV() error {
	fmt.Println("ewt sip scan av, ", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	directoryEntries, err := os.ReadDir(config.SIPLoc)
	if err != nil {
		return err
	}

	for _, entry := range directoryEntries {
		if entry.IsDir() && entry.Name() != "metadata" {
			fmt.Printf("  * Scanning %s for viruses\n", entry.Name())
			xfer := filepath.Join(config.SIPLoc, entry.Name())
			logName := filepath.Join(config.SIPLoc, "metadata", fmt.Sprintf("%s_clamscan.log", entry.Name()))
			if _, err := os.Create(logName); err != nil {
				return err
			}

			clamscanCmd := exec.Command("clamscan", "-r", xfer)
			cmdOut, err := clamscanCmd.CombinedOutput()
			if err != nil {
				return err
			}

			if err := os.WriteFile(logName, cmdOut, 0644); err != nil {
				return err
			}

		}
	}
	return nil
}
