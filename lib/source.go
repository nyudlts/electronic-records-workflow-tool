package lib

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func TransferSource() error {

	fmt.Println("ewt source transfer, version", VERSION)

	if err := loadConfig(); err != nil {
		return err
	}

	//create the logfile
	logFileName := filepath.Join(config.LogLoc, "rsync", fmt.Sprintf("%s-source-transfer-rsync.txt", config.CollectionCode))
	logFile, err := os.Create(logFileName)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(logFile)
	defer logFile.Close()

	fmt.Printf("  * Transferring %s to sip directory\n", config.SourceLoc)
	cmd := exec.Command("rsync", "-rav", config.SourceLoc, "sip")

	b, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	if _, err := writer.Write(b); err != nil {
		return err
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	if err := checkMDDir(); err != nil {
		panic(err)
	}

	fmt.Printf("  * Transfer complete\n")

	return nil
}

func checkMDDir() error {
	mdDir := filepath.Join(config.SIPLoc, "metadata")
	if _, err := os.Stat(mdDir); os.IsNotExist(err) {
		fmt.Printf("  * Creating `metadata` directory in SIP\n")
		if err := os.Mkdir(mdDir, 0755); err != nil {
			return err
		}
	}
	return nil
}

func PrintSourcePackageSize(directories bool) error {
	fmt.Println("ewt source size, version", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	if err := getPackageSize(config.SourceLoc); err != nil {
		return err
	}

	if directories {
		if err := printDirectoryStats(config.SourceLoc); err != nil {
			return err
		}
	}

	return nil
}
