package lib

import (
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

	//open/create a log file
	logFile, err := os.OpenFile(GetLog(RSTAR_PREP_PACKAGES), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//check aip_queue for files
	aipFiles, err := os.ReadDir(filepath.Join(config.WorkLoc, "in"))
	if err != nil {
		return fmt.Errorf("reading aip queue failed: %v", err)
	}

	if len(aipFiles) == 0 {
		fmt.Println("  * no packages found to process")
		log.Println("[INFO] no packages found to process")
		return nil
	}

	success := 0
	failure := 0
	for _, aipFile := range aipFiles {
		aipPath := filepath.Join(config.WorkLoc, "in", aipFile.Name())
		aipPathBytes, err := os.ReadFile(aipPath)
		if err != nil {
			fmt.Printf("  * reading aip file %s failed\n", aipFile.Name())
			log.Printf("[ERROR] reading aip file %s failed: %v", aipFile.Name(), err)
			failure++
			if err := moveToFailed(aipPath); err != nil {
				log.Printf("[ERROR] moving aip file %s to failed directory: %v", aipFile.Name(), err)
			}
			continue
		}

		aipLocation := strings.TrimSpace(string(aipPathBytes))
		if aipLocation == "" {
			fmt.Printf("  * aip file %s is empty\n", aipFile.Name())
			log.Printf("[ERROR] aip file %s is empty", aipFile.Name())
			failure++
			if err := moveToFailed(aipPath); err != nil {
				log.Printf("[ERROR] moving aip file %s to failed directory: %v", aipFile.Name(), err)
			}
			continue
		}

		if runtime.GOOS == "windows" {
			aipLocation = strings.ReplaceAll(aipLocation, "/", "\\")
			aipLocation = strings.Replace(aipLocation, "\\mnt\\amatica\\AIPsStore", config.AIPStoreLoc, 1)
		}

		fi, err := os.Stat(aipLocation)
		if err != nil {
			fmt.Printf("  * aip package %s does not exist\n", aipLocation)
			log.Printf("[ERROR] aip package %s does not exist: %v", aipLocation, err)
			failure++
			if err := moveToFailed(aipPath); err != nil {
				log.Printf("[ERROR] moving aip file %s to failed directory: %v", aipFile.Name(), err)
			}
			continue
		}

		msg := fmt.Sprintf("  * processing %s", fi.Name())
		fmt.Println(msg)
		log.Println("[INFO]", msg)
		if err := prepAmaticaAIP(aipLocation); err != nil {
			fmt.Printf("  * preparing package %s failed\n", fi.Name())
			log.Printf("[ERROR] preparing package %s failed: %v", fi.Name(), err)
			failure++
			if err := moveToFailed(aipPath); err != nil {
				log.Printf("[ERROR] moving aip file %s to failed directory: %v", aipFile.Name(), err)
			}
			continue
		}

		if err := moveToComplete(aipPath); err != nil {
			log.Printf("[ERROR] moving aip file %s to complete directory: %v", aipFile.Name(), err)
		}
		success++
	}

	count := success + failure
	fmt.Printf("\n  * rstar package prep complete, processed %d aip packages, %d failures\n", count, failure)
	return nil
}

func moveToFailed(aipPath string) error {
	failedDir := filepath.Join(config.WorkLoc, "failed")
	filename := filepath.Base(aipPath)
	failedPath := filepath.Join(failedDir, filename)
	if err := os.Rename(aipPath, failedPath); err != nil {
		return fmt.Errorf("could not move file to failed directory: %v", err)
	}
	return nil
}

func moveToComplete(aipPath string) error {
	completeDir := filepath.Join(config.WorkLoc, "complete")
	filename := filepath.Base(aipPath)
	completePath := filepath.Join(completeDir, filename)
	if err := os.Rename(aipPath, completePath); err != nil {
		return fmt.Errorf("Could not move file to complete directory: %v", err)
	}
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
	logFile, err := os.OpenFile(GetLog(RSTAR_VALIDATE), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}

	defer logFile.Close()
	log.SetOutput(logFile)

	//get aips to validate
	aips, err := os.ReadDir(filepath.Join(config.AIPLoc, "in"))
	if err != nil {
		return err
	}
	if len(aips) == 0 {
		fmt.Println("  * no aips found to validate")
		log.Println("[INFO] no aips found to validate")
		return nil
	}

	//validate each aip
	successCount := 0
	failureCount := 0
	for _, aip := range aips {
		if aip.IsDir() {

			erPath := filepath.Join(config.AIPLoc, "in", aip.Name())
			bag, err := bagit.GetExistingBag(erPath)
			if err != nil {
				log.Printf("Could not open bag: %v", err)
				failureCount++
				if err := moveAIP(aip.Name(), "in", "failed"); err != nil {
					log.Printf("[ERROR] moving aip %s to failed directory: %v", aip.Name(), err)
				}
				continue

			}

			if fullValidation {
				fmt.Printf("  * validating %s\n", aip.Name())
				log.Printf("[INFO] performing full validation on %s", aip.Name())
				if err := bag.ValidateBag(false, false); err != nil {
					failureCount++
					if err := moveAIP(aip.Name(), "in", "failed"); err != nil {
						log.Printf("[ERROR] moving aip %s to failed directory: %v", aip.Name(), err)
					}
					continue
				}
				fmt.Printf("  * validation complete for %s\n", aip.Name())
				log.Printf("[INFO] validation complete for %s", aip.Name())
			} else {
				fmt.Printf("  * fast validating %s\n", aip.Name())
				log.Printf("[INFO] performing fast validation on %s", aip.Name())
				if err := bag.ValidateBag(true, false); err != nil {
					log.Printf("[ERROR] fast validation failed for %s: %v", aip.Name(), err)
					failureCount++
					if err := moveAIP(aip.Name(), "in", "failed"); err != nil {
						log.Printf("[ERROR] moving aip %s to failed directory: %v", aip.Name(), err)
					}
					continue
				}
			}

		}
		if err := moveAIP(aip.Name(), "in", "valid"); err != nil {
			log.Printf("[ERROR] moving aip %s to valid directory: %v", aip.Name(), err)
		}
		fmt.Printf("  * validation complete for %s\n", aip.Name())
		log.Printf("[INFO] validation complete for %s", aip.Name())
		successCount++
	}
	fmt.Printf("  * rstar validation complete, %d successes, %d failures\n", successCount, failureCount)
	return nil
}

func moveAIP(aipName string, sourceDir string, destDir string) error {
	sourcePath := filepath.Join(config.AIPLoc, sourceDir, aipName)
	destPath := filepath.Join(config.AIPLoc, destDir, aipName)
	if err := os.Rename(sourcePath, destPath); err != nil {
		return fmt.Errorf("could not move file from %s to %s: %v", sourceDir, destDir, err)
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
	aips, err := os.ReadDir(filepath.Join(config.AIPLoc, "valid"))
	if err != nil {
		return err
	}
	if len(aips) == 0 {
		fmt.Println("  * no aips found to transfer")
		return nil
	}

	//create the log file
	xferLog, err := os.OpenFile(GetLog(RSTAR_TRANSFER), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0755)
	if err != nil {
		return err
	}
	defer xferLog.Close()

	//transfer aips
	failureCount := 0
	successCount := 0

	for _, aip := range aips {
		fmt.Printf("  * transferring %s\n", aip.Name())
		xferBag := filepath.Join(config.AIPLoc, "valid", aip.Name())
		xferCmd := exec.Command("rstar-scp.exp", xferBag)
		cmdOutput, err := xferCmd.CombinedOutput()
		if err != nil {
			failureCount++
			fmt.Printf("    * transfer of %s failed\n", aip.Name())
			log.Printf("[ERROR] transfer of %s failed: %v", aip.Name(), err)
			log.Printf("[ERROR] rstar-scp output: %s", string(cmdOutput))
			if err := moveAIP(aip.Name(), "valid", "failed"); err != nil {
				log.Printf("[ERROR] moving aip %s to failed directory: %v", aip.Name(), err)
			}
			continue
		}
		cmdOutput = append(cmdOutput, []byte("\n")...)

		if _, err = xferLog.Write(cmdOutput); err != nil {
			failureCount++
			fmt.Printf("    * writing rstar-scp output for %s failed\n", aip.Name())
			log.Printf("[ERROR] writing rstar-scp output for %s failed: %v", aip.Name(), err)
			if err := moveAIP(aip.Name(), "valid", "failed"); err != nil {
				log.Printf("[ERROR] moving aip %s to failed directory: %v", aip.Name(), err)
			}
			continue
		}

		if err := moveAIP(aip.Name(), "valid", "complete"); err != nil {
			log.Printf("[ERROR] moving aip %s to complete directory: %v", aip.Name(), err)
			continue
		}
		successCount++
	}

	fmt.Printf("  * rstar transfers complete, %d successes, %d failures\n", successCount, failureCount)

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
	aipStageLoc := filepath.Join(config.AIPLoc, "in", fi.Name())
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
		if !strings.HasSuffix(amaticaAIPLocation, "/") {
			amaticaAIPLocation = amaticaAIPLocation + "/"
		}
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
