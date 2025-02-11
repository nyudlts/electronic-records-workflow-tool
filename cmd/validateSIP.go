package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	rootCmd.AddCommand(validateCmd)
}

var workOrder aspace.WorkOrder
var row aspace.WorkOrderRow

var validateCmd = &cobra.Command{
	Use:   "validate-sip",
	Short: "validate sips prior to transfer to archivematica",
	Run: func(cmd *cobra.Command, args []string) {

		//load the project configuration
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		//create a logger
		logFile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s-adoc-validate-sip.log", adocConfig.CollectionCode)))
		if err != nil {
			panic(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)

		fmt.Printf("adoc validate-sip %s\n", version)
		log.Printf("[INFO] adoc-process validate-sip %s\n", version)
		fmt.Printf("* validating transfer package at %s\n", adocConfig.StagingLoc)
		log.Printf("[INFO] validating transfer package at %s\n", adocConfig.StagingLoc)

		if err := validate(); err != nil {
			panic(err)
		}
		fmt.Printf("* Report file written to %s", logFile.Name())
	},
}

func validate() error {
	//check that the source directory exists
	fmt.Print("  1. checking that source location exists: ")
	fileInfo, err := os.Stat(adocConfig.StagingLoc)
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
		fmt.Printf("source location %s does not exist", adocConfig.StagingLoc)
	} else {

		if !fileInfo.IsDir() {
			log.Printf("[ERROR] %s is not a directory\n", adocConfig.StagingLoc)
			fmt.Printf("source location %s is not a directory", adocConfig.StagingLoc)
		} else {
			log.Printf("[INFO] check 1. %s exists and is a directory\n", adocConfig.StagingLoc)
			fmt.Println("OK")
		}
	}

	//check that there is a metadata directory
	fmt.Print("  2. checking that source directory contains a metadata directory: ")
	mdDirLocation := filepath.Join(adocConfig.StagingLoc, "metadata")
	mdDir, err := os.Stat(mdDirLocation)
	if err != nil {
		fmt.Printf("source location %s does not contain a metadata directory", mdDirLocation)
		log.Printf("[ERROR] %s does not contain a metadata directory\n", adocConfig.StagingLoc)
	} else {

		if !mdDir.IsDir() {
			fmt.Printf("%s metadata directory is not a directory\n", mdDirLocation)
			log.Printf("[ERROR] %s is not a direcotry\n", mdDirLocation)
		} else {
			log.Printf("[INFO] check 2. %s contains a metadata directory\n", adocConfig.StagingLoc)
			fmt.Println("OK")
		}
	}

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

	//check that a transfer info exists
	fmt.Printf("  4. checking that %s contains a valid transfer-info.txt: ", mdDirLocation)
	xferInfoLocation := filepath.Join(mdDirLocation, "transfer-info.txt")
	_, err = os.Stat(xferInfoLocation)
	if err != nil {
		fmt.Println("transfer-info.txt does not exist in metadata directory\n")
		log.Println("[ERROR] transfer-info,txt does not exist in metadata directory")
	} else {
		xferBytes, err := os.ReadFile(xferInfoLocation)
		if err != nil {
			fmt.Printf("could not read transfer-info.txt: %s\n", xferInfoLocation)
			log.Printf("[ERROR]could not read transfer-info.txt: %s\n", xferInfoLocation)
		} else {
			transferInfo := TransferInfo{}
			if err := yaml.Unmarshal(xferBytes, &transferInfo); err != nil {
				fmt.Println("could unmarshal transfer-info.txt")
				log.Println("[ERROR] could not unmarshal transfer-info.txt")
			} else {
				if err := transferInfo.Validate(); err != nil {
					fmt.Printf("transfer-info.txt is not valid: %s\n", err.Error())
					log.Printf("[ERROR] transfer-info.txt is not valid: %s\n", err.Error())
				} else {
					log.Printf("[INFO] check 4. %s contains a valid transfer-info.txt \n", mdDirLocation)
					fmt.Println("OK")
				}
			}
		}
	}

	//get a list of componentIDs from work order
	fmt.Printf("  5. checking workorder %s for duplicate cuids: ", workorderName)
	componentIDs := []string{}
	//get an array of componentIDs
	dupeCount := 0
	for _, row := range workOrder.Rows {
		if contains(row.GetComponentID(), componentIDs) {
			log.Printf("[ERROR] duplicate componentID, %s, found in workorder\n", row.GetComponentID())
			dupeCount++
		} else {
			componentIDs = append(componentIDs, row.GetComponentID())
		}
	}
	sort.Strings(componentIDs)
	log.Printf("[INFO] check 5. %s contains %d duplicate cuids \n", workorderName, dupeCount)
	if dupeCount > 0 {
		fmt.Println("ERROR")
	} else {
		fmt.Println("OK")
	}

	fmt.Print("  6. checking all ER directories in workorder exist: ")
	missingDirs := 0
	for _, componentID := range componentIDs {
		erLocation := filepath.Join(adocConfig.StagingLoc, componentID)
		if _, err := os.Stat(erLocation); err != nil {
			missingDirs++
			log.Printf("[ERROR] componentID, %s is missing in transfered directories\n", componentID)
			//fmt.Printf("  * cuid %s is missing from transferred directories", componentID)
		}
	}
	log.Printf("[INFO] check 6. %s contains %d missing transfer directories \n", workorderName, missingDirs)

	if missingDirs > 0 {
		fmt.Println("ERROR")
	} else {
		fmt.Println("OK")
	}

	//check there are no extra directories in source location
	fmt.Print("  7. checking that there no extra directories or files in source location: ")
	sourceDirs, err := os.ReadDir(adocConfig.StagingLoc)
	if err != nil {
		log.Printf("[ERROR] duplicate componentID, %s, found in workorder", row.GetComponentID())
		fmt.Printf("[ERROR] duplicate componentID, %s, found in workorder\n", row.GetComponentID())
	} else {

		extraDirs := 0
		for _, sourceDir := range sourceDirs {
			if sourceDir.Name() != "metadata" {
				if !contains(sourceDir.Name(), componentIDs) {
					extraDirs++
					log.Printf("[ERROR] %s is not listed on workorder\n", sourceDir.Name())
				}
			}
		}

		log.Printf("[INFO] check 7. %s contained %d extra objects\n", adocConfig.StagingLoc, extraDirs)
		if extraDirs > 0 {
			fmt.Println("ERROR")
		} else {
			fmt.Println("OK")
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

		clamInfectedPtn := regexp.MustCompile("\nInfected files: 0")
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

		log.Printf("[INFO] check 8. %s contained %d failed clamscan scans", adocConfig.StagingLoc, failedClamScans)

		if failedClamScans > 0 {
			fmt.Println("ERROR")
		} else {
			fmt.Println("OK")
		}
	}

	return nil
}

func contains(s string, sl []string) bool {
	for _, sls := range sl {
		if s == sls {
			return true
		}
	}
	return false
}
