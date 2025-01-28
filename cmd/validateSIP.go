package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	rootCmd.AddCommand(validateCmd)
}

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
	fmt.Print("  1. checking that source location exists and is a directory: ")
	fileInfo, err := os.Stat(adocConfig.StagingLoc)
	if err != nil {
		log.Printf("[ERROR] %s\n", err.Error())
		return err
	}

	if !fileInfo.IsDir() {
		log.Printf("[ERROR] %s is not a directory\n", adocConfig.StagingLoc)
		return fmt.Errorf("source location is not a directory")
	}
	fmt.Println("OK")
	log.Printf("[INFO] check 1. %s exists and is a directory\n", adocConfig.StagingLoc)

	//check that there is a metadata directory
	fmt.Print("  2. checking that source directory contains a metadata directory: ")
	mdDirLocation := filepath.Join(adocConfig.StagingLoc, "metadata")
	mdDir, err := os.Stat(mdDirLocation)
	if err != nil {
		return err
	}

	if !mdDir.IsDir() {
		return fmt.Errorf("source metadata location is not a directory")
	}

	log.Printf("[INFO] check 2. %s contains a metadata directory\n", adocConfig.StagingLoc)
	fmt.Println("OK")

	//check that a workOrder exists
	fmt.Print("  3. checking that a valid workorder file exists: ")
	workorderName, err := getWorkOrderFile(mdDirLocation)
	if err != nil {
		return err
	}

	//check that the workorder is valid
	workOrder, err := parseWorkOrder(mdDirLocation, workorderName)
	if err != nil {
		return err
	}
	fmt.Println("OK")
	log.Printf("[INFO] check 3. %s contains a valid worker order \n", mdDirLocation)

	//check that a transfer info exists
	fmt.Printf("  4. checking that %s contains a valid transfer-info.txt: ", mdDirLocation)
	xferInfoLocation := filepath.Join(mdDirLocation, "transfer-info.txt")
	_, err = os.Stat(xferInfoLocation)
	if err != nil {
		return err
	}

	//validate transfer-info.txt
	xferBytes, err := os.ReadFile(xferInfoLocation)
	if err != nil {
		return err
	}
	transferInfo := TransferInfo{}
	if err := yaml.Unmarshal(xferBytes, &transferInfo); err != nil {
		return err
	}

	//validate transfer-info.txt
	if err := transferInfo.Validate(); err != nil {
		return err
	}
	log.Printf("[INFO] check 4. %s contains a valid transfer-info.txt \n", mdDirLocation)
	fmt.Println("OK")

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
		erLocation := filepath.Join(stagingLoc, componentID)
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
	sourceDirs, err := os.ReadDir(stagingLoc)
	if err != nil {
		panic(err)
	}

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

	//check that clamscan logs
	fmt.Print("  8. checking clamscan.logs: ")
	clamscanLogPtn := regexp.MustCompile("clamscan.log$")

	//check there are no failed clamscan logs
	mdFiles, err := os.ReadDir(mdDirLocation)
	if err != nil {
		return err
	}

	clamInfectedPtn := regexp.MustCompile("\nInfected files: 0")
	failedClamScans := 0
	for _, mdFile := range mdFiles {
		if clamscanLogPtn.MatchString(mdFile.Name()) {
			fileBytes, err := os.ReadFile(filepath.Join(mdDirLocation, mdFile.Name()))
			if err != nil {
				return err
			}
			if !clamInfectedPtn.Match(fileBytes) {
				failedClamScans++
				log.Printf("[ERROR] %s reports infected files\n", mdFile.Name())
			}
		}
	}

	log.Printf("[INFO] check 8. %s contained %d failed clamscan scans", adocConfig.StagingLoc, failedClamScans)

	if failedClamScans > 0 {
		fmt.Println("ERROR")
	} else {
		fmt.Println("OK")
	}

	//validation complete
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
