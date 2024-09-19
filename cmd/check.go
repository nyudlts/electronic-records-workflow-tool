package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/spf13/cobra"
)

func init() {
	checkCmd.Flags().StringVar(&sourceLoc, "source-location", "", "")
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use: "check",
	Run: func(cmd *cobra.Command, args []string) {
		if err := check(); err != nil {
			panic(err)
		} else {
			fmt.Println("All checks passed")
		}
	},
}

func check() error {
	fmt.Printf("aspace-preprocess v%s\n", version)

	//check that the source directory exists
	fmt.Print("  * checking that source location exists and is a directory: ")
	fileInfo, err := os.Stat(sourceLoc)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("source location is not a directory")
	}

	fmt.Println("OK")

	//check that there is a metadata directory
	fmt.Print("  * checking that source directory contains a metadata directory: ")
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
	fmt.Print("  * checking workorder file exists: ")
	workorderName, err := getWorkOrderFile(mdDirLocation)
	if err != nil {
		return err
	}
	fmt.Println("OK")

	//check that the workorder is valid
	fmt.Print("  * checking workorder is valid: ")
	workOrder, err := parseWorkOrder(mdDirLocation, workorderName)
	if err != nil {
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

	fmt.Println("  * checking ER directories exists")
	for _, componentID := range componentIDs {
		erLocation := filepath.Join(sourceLoc, componentID)
		if _, err := os.Stat(erLocation); err != nil {
			return err
		}
		fmt.Printf("    * %s exists\n", componentID)
	}

	//check there are no extra directories in source location
	fmt.Print("  * checking that no extra directories are source location: ")
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
