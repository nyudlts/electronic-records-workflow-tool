package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	bagit "github.com/nyudlts/go-bagit"
)

var (
	woMatcher = regexp.MustCompile("aspace_wo.tsv$")
	tiMatcher = regexp.MustCompile("transfer-info.txt")
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
	count := 0
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

		//new line if not the first package
		if count > 0 {
			fmt.Println()
		}
		count++
		msg := fmt.Sprintf("  * processing %s", fi.Name())
		fmt.Println(msg)
		log.Println("[INFO]", msg)
		if err := prepAmaticaAIP(aipLocation); err != nil {
			log.Printf("[ERROR] preparing package %s failed: %v", fi.Name(), err)
			return fmt.Errorf("preparing package %s failed: %v", fi.Name(), err)
		}
	}
	fmt.Printf("\n  * rstar package prep complete, processed %d packages\n", count)
	return nil
}

func PrepareSinglePackage(aipLocation string) error {
	fmt.Println("ewt rstar prep single package", VERSION)

	//load the config
	if err := loadConfig(); err != nil {
		return err
	}

	//create a log file
	logFile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s-rstar-prep-package.log", config.CollectionCode)))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//check the aip location
	fi, err := os.Stat(aipLocation)
	if err != nil {
		return fmt.Errorf("aip package %s does not exist: %v", aipLocation, err)
	}

	msg := fmt.Sprintf("  * processing %s", fi.Name())
	fmt.Println(msg)
	log.Println("[INFO]", msg)

	if err := prepAmaticaAIP(aipLocation); err != nil {
		return err
	}

	fmt.Println("\n  * rstar package prep complete")
	return nil
}

func ValidateRStarPackages(fullValidation bool) error {
	fmt.Println("ewt rstar validate", VERSION)

	//load the config
	if err := loadConfig(); err != nil {
		return err
	}

	//create a log file
	logFile, err := os.Create(fmt.Sprintf("logs/%s-rstar-validate.log", config.CollectionCode))
	if err != nil {
		return err
	}

	defer logFile.Close()
	log.SetOutput(logFile)

	//get aips to validate
	aips, err := os.ReadDir(config.AIPLoc)
	if err != nil {
		return err
	}
	if len(aips) == 0 {
		fmt.Println("  * no aips found to validate")
		log.Println("[INFO] no aips found to validate")
		return nil
	}

	//validate each aip
	for _, aip := range aips {
		if aip.IsDir() {

			erPath := filepath.Join(config.AIPLoc, aip.Name())
			bag, err := bagit.GetExistingBag(erPath)
			if err != nil {
				return err
			}

			if fullValidation {
				fmt.Printf("  * validating %s\n", aip.Name())
				log.Printf("[INFO] performing full validation on %s", aip.Name())
				if err := bag.ValidateBag(false, false); err != nil {
					log.Printf("[ERROR] full validation failed for %s: %v", aip.Name(), err)
					return err
				}
				fmt.Printf("  * validation complete for %s\n", aip.Name())
				log.Printf("[INFO] validation complete for %s", aip.Name())
			} else {
				fmt.Printf("  * fast validating %s\n", aip.Name())
				log.Printf("[INFO] performing fast validation on %s", aip.Name())
				if err := bag.ValidateBag(true, false); err != nil {
					log.Printf("[ERROR] fast validation failed for %s: %v", aip.Name(), err)
					return err
				}
				fmt.Printf("  * validation complete for %s\n", aip.Name())
				log.Printf("[INFO] validation complete for %s", aip.Name())
			}

		}
	}
	return nil
}

func TransferRStarPackages() error {
	fmt.Println("ewt rstar transfer", VERSION)
	//load the config
	if err := loadConfig(); err != nil {
		return err
	}

	//read aip directory
	aips, err := os.ReadDir(config.AIPLoc)
	if err != nil {
		return err
	}
	if len(aips) == 0 {
		fmt.Println("  * no aips found to transfer")
		return nil
	}

	//create the log file
	xferLogFile := filepath.Join("logs", fmt.Sprintf("%s-rstar-transfer.txt", config.CollectionCode))
	_, err = os.Create(xferLogFile)
	if err != nil {
		return err
	}

	//transfer aips
	for _, aip := range aips {
		fmt.Printf("  * transferring %s\n", aip.Name())
		xferBag := filepath.Join(config.AIPLoc, aip.Name())
		xferCmd := exec.Command("rstar-scp.exp", xferBag)
		cmdOutput, err := xferCmd.CombinedOutput()
		if err != nil {
			return err
		}
		cmdOutput = append(cmdOutput, []byte("\n")...)

		xferLog, err := os.OpenFile(xferLogFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0775)
		if err != nil {
			return err
		}
		defer xferLog.Close()

		if _, err = xferLog.Write(cmdOutput); err != nil {
			return err
		}
	}

	fmt.Println("  * rstar transfer complete")

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

	if len(objs) == 0 {
		fmt.Println("  * aip directory is already clean")
		return nil
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
	msg := "copying package to aip directory"
	fmt.Printf("    * %s\n", msg)
	log.Printf("[INFO] %s + %s", msg, fi.Name())
	var cmd *exec.Cmd
	var out []byte
	if runtime.GOOS == "windows" {
		cmd = exec.Command("robocopy", amaticaAIPLocation, aipStageLoc, "/E", "/DCOPY:DAT")
		out, err = cmd.CombinedOutput()
		if err != nil && err.Error() != "exit status 1" {
			return err
		}
	} else {
		cmd = exec.Command("rsync", "-rav", amaticaAIPLocation, config.AIPLoc)
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

	msg = ("Updating package")
	fmt.Println("    * " + msg)
	log.Printf("[INFO] %s %s", msg, fi.Name())

	if err := updatePackage(aipStageLoc); err != nil {
		return err
	}

	fmt.Println("  * processing complete for ", fi.Name())

	return nil
}

func updatePackage(bagLocation string) error {

	fmt.Println("      * opening bag")
	bag, err := bagit.GetExistingBag(bagLocation)
	if err != nil {
		return err
	}

	if runtime.GOOS == "linux" {
		//validate the bag
		fmt.Println("      * Validating bag")
		if err := bag.ValidateBag(false, false); err != nil {
			return err
		}
	}

	//locate the work order
	fmt.Println("      * Locating work order")
	matches := bag.Payload.FindFilesInPayload(woMatcher)
	if len(matches) != 1 {
		return fmt.Errorf("no workorder found")
	}
	woPath := matches[0].Path

	//move the work order to the bag's root
	fmt.Println("      * Moving work order to bag's root")
	if err := bag.AddFileToBagRoot(woPath); err != nil { // this is not returning an err it is panicing fix in go-bagit
		log.Printf("[WARNING] moving work order to bag's root failed: %v", err)
	}

	//locate the transfer-info.txt
	fmt.Println("      * Locating transfer-info.txt")
	matches = bag.Payload.FindFilesInPayload(tiMatcher)
	if len(matches) != 1 {
		return fmt.Errorf("no transfer-info.txt found")
	}
	tiPath := matches[0].Path
	tiPath = strings.ReplaceAll(tiPath+"/", bagLocation, "")

	//create a tag set from transfer-info.txt
	fmt.Println("      * Creating new tag set from transfer-info.txt")
	transferInfo, err := bagit.NewTagSet(tiPath, bagLocation)
	if err != nil {
		return err
	}

	//Update the hostname
	fmt.Println("      * Adding hostname to tag set")
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	transferInfo.Tags["nyu-dl-hostname"] = hostname

	//add pathname to the tag-set
	fmt.Println("      * Adding bag's path to tag set")
	path, err := filepath.Abs(bagLocation)
	if err != nil {
		return err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	transferInfo.Tags["nyu-dl-pathname"] = path

	//backup bag-info
	fmt.Println("      * Backing up bag-info.txt")
	bagInfoLocation := filepath.Join(bagLocation, "bag-info.txt")
	backupLocation := filepath.Join(config.ProjectLoc, ".bag-info.txt")

	biBytes, err := os.ReadFile(bagInfoLocation)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupLocation, biBytes, 0777); err != nil {
		return err
	}

	//getting tagset from bag-info
	fmt.Println("      * Creating new tag set from bag-info.txt")
	bagInfo, err := bagit.NewTagSet("bag-info.txt", bagLocation)
	if err != nil {
		return err
	}

	//merge tagsets
	fmt.Println("      * Merging Tag Sets")
	bagInfo.AddTags(transferInfo.Tags)
	bagInfoBytes := bagInfo.GetTagSetAsByteSlice()

	//write new bag-info.txt
	fmt.Println("      * Rewriting bag-info.txt")
	if err := os.WriteFile(bagInfoLocation, bagInfoBytes, 0777); err != nil {
		return err
	}

	//create new tag manifest
	fmt.Println("      * Creating new tagmanifest-sha256.txt")
	tagManifest, err := bagit.NewManifest(bagLocation, "tagmanifest-sha256.txt")
	if err != nil {
		return err
	}

	//update the checksum for bag-info.txt
	fmt.Println("      * Updating checksum for bag-info.txt in tagmanifest-sha256.txt")
	if err := tagManifest.UpdateManifest("bag-info.txt"); err != nil {
		return err
	}

	fmt.Println("      * Rewriting tagmanifest-sha256.txt")
	if err := tagManifest.Serialize(); err != nil {
		return err
	}

	if runtime.GOOS == "linux" {
		//validate the updated bag
		fmt.Println("      * Validating the updated bag")
		if err := bag.ValidateBag(false, false); err != nil {
			return err
		}
	}

	//delete the backup bag-info
	fmt.Println("      * Deleting backup bag-info.txt")
	if err := os.Remove(backupLocation); err != nil {
		return err
	}

	fmt.Println("    * Package update complete")

	return nil
}
