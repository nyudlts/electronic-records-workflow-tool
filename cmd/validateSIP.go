package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	validateCmd.Flags().StringVar(&stagingLoc, "staging-location", "", "location of sip to validate (required)")
	rootCmd.AddCommand(validateCmd)
}

var writer *bufio.Writer
var report *os.File

var validateCmd = &cobra.Command{
	Use:   "validate-sip",
	Short: "validate sips prior to transfer to archivematica",
	Run: func(cmd *cobra.Command, args []string) {
		getWriter()
		fmt.Printf("adoc-process validate v%s\n", version)
		writer.WriteString(fmt.Sprintf("[INFO] adoc-process validate v%s\n", version))
		fmt.Printf("* validating transfer package at %s\n", stagingLoc)
		writer.WriteString(fmt.Sprintf("[INFO] validating transfer package at %s\n", stagingLoc))

		if err := validate(); err != nil {
			panic(err)
		}
		fmt.Printf("* Report file written to %s", report.Name())
		writer.Flush()
	},
}

func validate() error {
	//check that the source directory exists
	fmt.Print("  1. checking that source location exists and is a directory: ")
	fileInfo, err := os.Stat(stagingLoc)
	if err != nil {
		writer.WriteString(fmt.Sprintf("[ERROR] %s\n", err.Error()))
		return err
	}

	if !fileInfo.IsDir() {
		writer.WriteString(fmt.Sprintf("[ERROR] %s is not a directory\n", stagingLoc))
		return fmt.Errorf("source location is not a directory")
	}
	fmt.Println("OK")
	writer.WriteString(fmt.Sprintf("[INFO] check 1. %s exists and is a directory\n", stagingLoc))

	//check that there is a metadata directory
	fmt.Print("  2. checking that source directory contains a metadata directory: ")
	mdDirLocation := filepath.Join(stagingLoc, "metadata")
	mdDir, err := os.Stat(mdDirLocation)
	if err != nil {
		return err
	}

	if !mdDir.IsDir() {
		return fmt.Errorf("source metadata location is not a directory")
	}

	writer.WriteString(fmt.Sprintf("[INFO] check 2. %s contains a metadata directory\n", stagingLoc))
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
	writer.WriteString(fmt.Sprintf("[INFO] check 3. %s contains a valid worker order \n", mdDirLocation))

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
	writer.WriteString(fmt.Sprintf("[INFO] check 4. %s contains a valid transfer-info.txt \n", mdDirLocation))
	fmt.Println("OK")

	//get a list of componentIDs from work order
	fmt.Printf("  5. checking workorder %s for duplicate cuids: ", workorderName)
	componentIDs := []string{}
	//get an array of componentIDs
	dupeCount := 0
	for _, row := range workOrder.Rows {
		if contains(row.GetComponentID(), componentIDs) {
			writer.WriteString(fmt.Sprintf("[ERROR] duplicate componentID, %s, found in workorder\n", row.GetComponentID()))
			//fmt.Printf("  * duplicate cuid found: %s", row.GetComponentID())
			dupeCount++
		} else {
			componentIDs = append(componentIDs, row.GetComponentID())
		}
	}
	sort.Strings(componentIDs)
	writer.WriteString(fmt.Sprintf("[INFO] check 5. %s contains %d duplicate cuids \n", workorderName, dupeCount))
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
			writer.WriteString(fmt.Sprintf("[ERROR] componentID, %s is missing in transfered directories\n", componentID))
			//fmt.Printf("  * cuid %s is missing from transferred directories", componentID)
		}
	}
	writer.WriteString(fmt.Sprintf("[INFO] check 6. %s contains %d missing transfer directories \n", workorderName, missingDirs))

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
				writer.WriteString(fmt.Sprintf("[ERROR] %s is not listed on workorder\n", sourceDir.Name()))
				//fmt.Printf("\t%s is not listed on workorder\n", sourceDir.Name())
			}
		}
	}

	writer.WriteString(fmt.Sprintf("[INFO] check 7. %s contained %d extra objects\n", stagingLoc, extraDirs))
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
				writer.WriteString(fmt.Sprintf("[ERROR] %s reports infected files\n", mdFile.Name()))
				//fmt.Printf("%s reports infected files", mdFile.Name())
			}
		}
	}

	writer.WriteString(fmt.Sprintf("[INFO] check 8. %s contained %d failed clamscan scans", stagingLoc, failedClamScans))

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

func getWriter() error {
	_, dirName := filepath.Split(stagingLoc)
	var err error
	report, err = os.Create(fmt.Sprintf("adoc-validation-report-%s.txt", dirName))
	if err != nil {
		return err
	}
	writer = bufio.NewWriter(report)
	return nil
}
