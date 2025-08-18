package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func TransferSource() error {

	fmt.Println("ewt source transfer, version", VERSION)

	if err := loadConfig(); err != nil {
		return err
	}

	//create the rsync/robocopy output file
	logFileName := filepath.Join(config.LogLoc, "rsync", fmt.Sprintf("%s-source-transfer-rsync.txt", config.CollectionCode))
	logFile, err := os.Create(logFileName)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(logFile)
	defer logFile.Close()

	fmt.Printf("  * Transferring %s to sip directory\n", config.SourceLoc)
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("robocopy", config.SourceLoc, config.SIPLoc, "/E", "/DCOPY:DAT")
	} else {
		cmd = exec.Command("rsync", "-rav", config.SourceLoc, config.SIPLoc)
	}

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

	//check if the metadata director exists
	mdDirLoc := filepath.Join(config.SIPLoc, "metadata")
	if _, err := os.Stat(mdDirLoc); err != nil {
		log.Printf("  * creating metadata directory in %s\n", config.SIPLoc)
		if err := os.Mkdir(mdDirLoc, 0755); err != nil {
			return err
		}
		fmt.Printf("  * created metadata directory in %s", config.SIPLoc)
	}

	fmt.Println("  * Transfer complete")
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
