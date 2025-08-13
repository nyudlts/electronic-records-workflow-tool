package lib

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

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
