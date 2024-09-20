package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	checkCmd.Flags().StringVar(&sourceLoc, "source-location", "", "")
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use: "check",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("aspace-preprocess v%s\n", version)
		fmt.Printf("* validating transfer package at %s \n", sourceLoc)
		if err := check(); err != nil {
			panic(err)
		} else {
			fmt.Printf("* success, all checks passed for %s", sourceLoc)
		}
	},
}

func check() error {
	//check that the source directory exists
	fmt.Print("  1. checking that source location exists and is a directory: ")
	fileInfo, err := os.Stat(sourceLoc)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("source location is not a directory")
	}

	fmt.Println("OK")

	//check that there is a metadata directory
	fmt.Print("  2. checking that source directory contains a metadata directory: ")
	mdDirLocation := filepath.Join(sourceLoc, "metadata")
	mdDir, err := os.Stat(mdDirLocation)
	if err != nil {
		return err
	}

	if !mdDir.IsDir() {
		return fmt.Errorf("source metadata location is not a directory")
	}
	fmt.Println("OK")

	//check that a workOrder exists
	fmt.Print("  3. checking workorder file exists: ")
	workorderName, err := getWorkOrderFile(mdDirLocation)
	if err != nil {
		return err
	}
	fmt.Println("OK")

	//check that the workorder is valid
	fmt.Print("  4. validating workorder: ")
	workOrder, err := parseWorkOrder(mdDirLocation, workorderName)
	if err != nil {
		return err
	}
	fmt.Println("OK")

	//check that a transfer info exists
	fmt.Print("  5. checking transfer-info.txt exists: ")
	xferInfoLocation := filepath.Join(mdDirLocation, "transfer-info.txt")
	_, err = os.Stat(xferInfoLocation)
	if err != nil {
		return err
	}
	fmt.Println("OK")

	//parsing transfer-info.txt
	fmt.Print("  6. parsing transfer-info.tx: ")
	xferBytes, err := os.ReadFile(xferInfoLocation)
	if err != nil {
		return err
	}
	transferInfo := TransferInfo{}
	if err := yaml.Unmarshal(xferBytes, &transferInfo); err != nil {
		return err
	}

	fmt.Println("OK")
	//validate transfer-info.txt
	fmt.Print("  7. validating transfer-info.txt: ")
	if err := transferInfo.Validate(); err != nil {
		return err
	}
	fmt.Println("OK")

	//get a list of componentIDs from work order
	componentIDs := []string{}
	//get an array of componentIDs
	for _, row := range workOrder.Rows {
		if contains(row.GetComponentID(), componentIDs) {
			return fmt.Errorf("duplicate componentID, %s, found in workorder", row.GetComponentID())
		} else {
			componentIDs = append(componentIDs, row.GetComponentID())
		}
	}
	sort.Strings(componentIDs)

	fmt.Println("  8. checking ER directories exists")
	for _, componentID := range componentIDs {
		erLocation := filepath.Join(sourceLoc, componentID)
		if _, err := os.Stat(erLocation); err != nil {
			return err
		}
		fmt.Printf("    %s exists\n", componentID)
	}

	//check there are no extra directories in source location
	fmt.Print("  9. checking that there no extra directories or files in source location: ")
	sourceDirs, err := os.ReadDir(sourceLoc)
	if err != nil {
		return err
	}

	for _, sourceDir := range sourceDirs {
		if sourceDir.Name() != "metadata" {
			if !contains(sourceDir.Name(), componentIDs) {
				fmt.Println()
				return fmt.Errorf("%s is not listed on workorder", sourceDir.Name())
			}
		}
	}
	fmt.Println("OK")
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
