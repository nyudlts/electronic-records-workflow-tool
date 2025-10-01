package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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
	logFile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s-rstar-prep-packages.log", config.CollectionCode)))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//get the aip file
	aipFileLoc := filepath.Join(config.LogLoc, fmt.Sprintf("%s-aip-file.txt", config.CollectionCode))
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
		if runtime.GOOS == "windows" {
			aipLocation = strings.ReplaceAll(aipLocation, "/", "\\")
			aipLocation = strings.Replace(aipLocation, "\\mnt\\amatica\\AIPsStore", config.AIPStoreLoc, 1)
		}

		fi, err := os.Stat(aipLocation)
		if err != nil {
			log.Printf("[ERROR] aip package %s does not exist: %v", aipFileLoc, err)
			return fmt.Errorf("aip package %s does not exist: %v", aipFileLoc, err)
		}

		msg := fmt.Sprintf("  * processing %s", fi.Name())
		fmt.Println(msg)
		log.Println("[INFO]", msg)
		if err := prepAmaticaAIP(aipLocation); err != nil {
			log.Printf("[ERROR] preparing package %s failed: %v", fi.Name(), err)
			return fmt.Errorf("preparing package %s failed: %v", fi.Name(), err)
		}
	}
	fmt.Println("  *rstar package prep complete")
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

func CleanAIPDirectory() error {
	fmt.Println("ewt rstar clean", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	objs, err := os.ReadDir(config.AIPLoc)
	if err != nil {
		return err
	}

	for _, obj := range objs {
		objPath := filepath.Join(config.AIPLoc, obj.Name())
		fmt.Printf("  * removing %s\n", obj.Name())
		if err := os.RemoveAll(objPath); err != nil {
			return err
		}
	}
	return nil
}

func prepAmaticaAIP(amaticaAIPLocation string) error {
	//copy the package to aip directory
	fi, err := os.Stat(amaticaAIPLocation)
	if err != nil {
		return err
	}
	aipStageLoc := filepath.Join(config.AIPLoc, fi.Name())
	msg := fmt.Sprintf("copying %s to aip directory", fi.Name())
	fmt.Printf("  * %s\n", msg)
	log.Printf("[INFO] %s", msg)
	var cmd *exec.Cmd
	var out []byte
	if runtime.GOOS == "windows" {
		cmd = exec.Command("robocopy", amaticaAIPLocation, aipStageLoc, "/E", "/DCOPY:DAT")
		out, err = cmd.CombinedOutput()
		if err != nil && err.Error() != "exit status 1" {
			return err
		}
	} else {
		cmd = exec.Command("rsync", "-rav", amaticaAIPLocation, aipStageLoc)
		out, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println()
			return err
		}
	}

	if err := os.WriteFile(filepath.Join("logs", "rsync", fmt.Sprintf("%s-rsync-output.txt", fi.Name())), out, 0775); err != nil {
		fmt.Println()
		return err
	}

	msg = fmt.Sprintf("  * Updating package %s", fi.Name())
	fmt.Println(msg)
	log.Printf("[INFO] %s", msg)

	if err := updatePackage(aipStageLoc, config.ProjectLoc); err != nil {
		return err
	}

	return nil
}

func updatePackage(aipStageLoc, tmpLoc string) error {
	return nil
}
