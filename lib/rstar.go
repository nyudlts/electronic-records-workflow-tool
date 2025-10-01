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
	fmt.Println("  * rstar package prep complete")
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
	fmt.Printf("    * %s\n", msg)
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

	msg = fmt.Sprintf("    * Updating package %s", fi.Name())
	fmt.Println(msg)
	log.Printf("[INFO] %s", msg)

	if err := updatePackage(aipStageLoc); err != nil {
		return err
	}

	return nil
}

func updatePackage(bagLocation string) error {

	fmt.Println("      * opening bag", filepath.Base(bagLocation))
	bag, err := bagit.GetExistingBag(bagLocation)
	if err != nil {
		return err
	}

	//validate the bag
	fmt.Printf("      * Validating bag %s\n", filepath.Base(bagLocation))
	if err := bag.ValidateBag(false, false); err != nil {
		return err
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
	if err := bag.AddFileToBagRoot(woPath); err != nil {
		return err
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
	fmt.Println("      * Adding hostname to tag set: ")
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

	/*

		fmt.Printf("  * Opening bag-info.txt: ")
		bagInfoFile, err := os.Open(bagInfoLocation)
		if err != nil {
			return err
		}
		defer bagInfoFile.Close()
		fmt.Printf("OK\n")

		fmt.Printf("  * Rewriting bag-info.txt: ")
		if err := os.WriteFile(bagInfoLocation, bagInfoBytes, 0777); err != nil {
			return err
		}
		fmt.Printf("OK\n")

		//create new manifest object for tagmanifest-sha256.txt
		fmt.Printf("  * Creating new tagmanifest-sha256.txt: ")
		tagManifest, err := bagit.NewManifest(bagLocation, "tagmanifest-sha256.txt")
		if err != nil {
			return err
		}
		fmt.Printf("OK\n")

		//update the checksum for bag-info.txt
		fmt.Printf("  * Updating checksum for bag-info.txt in tagmanifest-sha256.txt: ")
		if err := tagManifest.UpdateManifest("bag-info.txt"); err != nil {
			return err
		}
		fmt.Printf("OK\n")

		fmt.Printf("  * Rewriting tagmanifest-sha256.txt: ")
		if err := tagManifest.Serialize(); err != nil {
			return err
		}
		fmt.Printf("OK\n")

		//validate the updated bag
		fmt.Printf("\nValidating the updated bag: ")
		if err := bag.ValidateBag(false, false); err != nil {
			return err
		}
		fmt.Printf("OK\n")

		//delete the backup bag-info
		fmt.Printf("Deleting backup bag-info.txt: ")
		if err := os.Remove(backupLocation); err != nil {
			return err
		}
		fmt.Printf("OK\n")


	*/
	fmt.Println("    * Package preparation complete")
	fmt.Println()
	return nil
}
