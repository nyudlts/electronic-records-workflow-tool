package lib

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/nyudlts/go-aspace"
	"gopkg.in/yaml.v2"
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

	var workOrder aspace.WorkOrder

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

	//check that a workOrder exists
	fmt.Print("  3. checking that a valid workorder file exists: ")
	workorderName, err := getWorkOrderFile(mdDirLocation)
	if err != nil {
		fmt.Printf("metadata directory %s does not contain a work order\n", mdDirLocation)
		log.Printf("[ERROR] metadata directory %s does not contain a work order\n", mdDirLocation)
	} else {
		//check that the workorder is valid
		workOrder, err = parseWorkOrder(mdDirLocation, workorderName)
		if err != nil {
			fmt.Printf("work order %s is not valid: %s\n", mdDirLocation, err.Error())
			log.Printf("[ERROR] work order %s is not valid: %s\n", mdDirLocation, err.Error())
		} else {
			fmt.Println("OK")
			log.Printf("[INFO] check 3. %s contains a valid worker order \n", mdDirLocation)
		}
	}

	//get a list of componentIDs from work order
	fmt.Printf("  4. checking workorder %s for duplicate cuids: ", workorderName)
	componentIDs := []string{}
	//get an array of componentIDs
	dupeCount := 0
	for _, row := range workOrder.Rows {
		if woContains(row.GetComponentID(), componentIDs) {
			log.Printf("[ERROR] duplicate componentID, %s, found in workorder\n", row.GetComponentID())
			dupeCount++
		} else {
			componentIDs = append(componentIDs, row.GetComponentID())
		}
	}

	sort.Strings(componentIDs)
	log.Printf("[INFO] check 4. %s contains %d duplicate cuids \n", workorderName, dupeCount)
	if dupeCount > 0 {
		fmt.Println("ERROR")
	} else {
		fmt.Println("OK")
	}

	//check that all componentIDs in the workorder exist in the SIP
	fmt.Print("  5. checking all ER directories in workorder exist: ")
	missingDirs := 0
	for _, componentID := range componentIDs {
		erLocation := filepath.Join(config.SIPLoc, componentID)
		if _, err := os.Stat(erLocation); err != nil {
			missingDirs++
			log.Printf("[ERROR] componentID, %s is missing in transfered directories\n", componentID)
			//fmt.Printf("  * cuid %s is missing from transferred directories", componentID)
		}
	}
	log.Printf("[INFO] check 5. %s contains %d missing transfer directories \n", workorderName, missingDirs)

	if missingDirs > 0 {
		fmt.Println("ERROR")
	} else {
		fmt.Println("OK")
	}

	//check there are no extra directories in source location
	fmt.Print("  6. checking that there no extra directories or files in SIP directory: ")
	sourceDirs, err := os.ReadDir(config.SIPLoc)
	if err != nil {
		log.Printf("[ERROR] could not read SIP directory %s: %s\n", config.SIPLoc, err.Error())
		fmt.Printf("[ERROR] could not read SIP directory %s: %s\n", config.SIPLoc, err.Error())
	} else {

		extraDirs := 0
		for _, sourceDir := range sourceDirs {
			if sourceDir.Name() != "metadata" {
				if !woContains(sourceDir.Name(), componentIDs) {
					extraDirs++
					log.Printf("[ERROR] %s is not listed on workorder\n", sourceDir.Name())
				}
			}
		}

		log.Printf("[INFO] check 6. %s contained %d extra objects\n", config.SIPLoc, extraDirs)
		if extraDirs > 0 {
			fmt.Println("ERROR")
		} else {
			fmt.Println("OK")
		}
	}

	//check that SIP contains a valid transfer-info.txt
	fmt.Print("  7. checking that valid transfer-info.txt exists: ")
	xferInfoLocation := filepath.Join(mdDirLocation, "transfer-info.txt")
	_, err = os.Stat(xferInfoLocation)
	if err != nil {
		fmt.Println("transfer-info.txt does not exist in metadata directory")
		log.Println("[ERROR] transfer-info,txt does not exist in metadata directory")
	} else {
		xferBytes, err := os.ReadFile(xferInfoLocation)
		if err != nil {
			fmt.Printf("could not read transfer-info.txt: %s\n", xferInfoLocation)
			log.Printf("[ERROR]could not read transfer-info.txt: %s\n", xferInfoLocation)
		} else {
			transferInfo := TransferInfo{}
			if err := yaml.Unmarshal(xferBytes, &transferInfo); err != nil {
				fmt.Println("could not unmarshal transfer-info.txt")
				log.Println("[ERROR] could not unmarshal transfer-info.txt")
			} else {
				if err := transferInfo.Validate(); err != nil {
					fmt.Printf("transfer-info.txt is not valid: %s\n", err.Error())
					log.Printf("[ERROR] transfer-info.txt is not valid: %s\n", err.Error())
				} else {
					log.Printf("[INFO] check 7. %s contains a valid transfer-info.txt \n", mdDirLocation)
					fmt.Println("OK")
				}
			}
		}
	}

	//check that clamscan logs
	fmt.Print("  8. checking clamscan.logs: ")
	clamscanLogPtn := regexp.MustCompile("clamscan.log$")

	//check there are no failed clamscan logs
	mdFiles, err := os.ReadDir(mdDirLocation)
	if err != nil {
		log.Printf("[ERROR] cannot open metadata directory: %s\n", mdDirLocation)
		fmt.Printf("cannot open metadata directory: %s\n", mdDirLocation)
	} else {
		failedClamScans := 0
		for _, mdFile := range mdFiles {
			if clamscanLogPtn.MatchString(mdFile.Name()) {
				fileBytes, err := os.ReadFile(filepath.Join(mdDirLocation, mdFile.Name()))
				if err != nil {
					log.Printf("[ERROR] cannot read clamscan log: %s", mdFile.Name())
					fmt.Printf("cannot read clamscan log: %s\n", mdFile.Name())
				} else {
					if !clamInfectedPtn.Match(fileBytes) {
						failedClamScans++
						log.Printf("[ERROR] clamscan %s contained infected files", mdFile.Name())
					}
				}
			}
		}

		log.Printf("[INFO] check 8. SIP contained %d failed clamscan scans", failedClamScans)

		if failedClamScans > 0 {
			fmt.Println("ERROR")
		} else {
			fmt.Println("OK")
		}
	}

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
			//this needs to work with clamdscan
			clamscanCmd := exec.Command("clamscan", "-r", xfer)

			cmdOut, err := clamscanCmd.CombinedOutput()
			if err != nil {
				return err
			}

			logName := filepath.Join(config.SIPLoc, "metadata", fmt.Sprintf("%s_clamscan.log", entry.Name()))
			if _, err := os.Create(logName); err != nil {
				return err
			}

			if err := os.WriteFile(logName, cmdOut, 0644); err != nil {
				return err
			}

		}
	}
	return nil
}

func woContains(s string, sl []string) bool {
	for _, sls := range sl {
		if s == sls {
			return true
		}
	}
	return false
}
