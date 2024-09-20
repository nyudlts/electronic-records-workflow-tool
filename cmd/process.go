package cmd

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const version = "0.2.0"

var (
	partnerCode  string
	resourceCode string
)

func init() {
	processCmd.Flags().StringVar(&sourceLoc, "source-location", "", "the location of the package to be transferred to r*")
	processCmd.Flags().StringVar(&stagingLoc, "staging-location", "", "the location of the staging location for Archivematica")
	rootCmd.AddCommand(processCmd)
}

var processCmd = &cobra.Command{
	Use: "process",
	Run: func(cmd *cobra.Command, args []string) {
		process()
	},
}

func process() {
	fmt.Printf("adoc-preprocess v%s\n", version)
	flag.Parse()
	params := Params{}

	logFile, err := os.Create("adoc-preprocess.log") //this should have the call number of the collection e.g. fales_mss318.log
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Println("[INFO] checking source directory exists")
	//check that source exists and is a Directory
	if err := isDirectory(sourceLoc); err != nil {
		panic(err)
	}
	params.Source = sourceLoc

	//check that staging location exists and is a Directory
	if err := isDirectory(stagingLoc); err != nil {
		panic(err)
	}
	params.StagingLoc = stagingLoc

	log.Println("[INFO] checking metadata directory exists")
	//check that metadata directory exists and is a directory
	mdDir := filepath.Join(sourceLoc, "metadata")
	if err := isDirectory(mdDir); err != nil {
		panic(err)
	}

	log.Println("[INFO] checking work order exists")

	//find a work order
	workorderName, err := getWorkOrderFile(mdDir)
	if err != nil {
		panic(err)
	}

	//getting partner and resource code
	log.Println("[INFO] getting partner and resource code")
	params.PartnerCode, params.ResourceCode = getPartnerAndResource(workorderName)

	//load the work order
	log.Println("[INFO] parsing work order")

	params.WorkOrder, err = parseWorkOrder(mdDir, workorderName)
	if err != nil {
		panic(err)
	}

	//find the transfer info file
	log.Println("[INFO] checking transfer info exists")
	transferInfoLoc := filepath.Join(mdDir, "transfer-info.txt")
	if _, err = os.Stat(transferInfoLoc); err != nil {
		panic(err)
	}

	//create the transfer-info struct
	log.Println("[INFO] parsing transfer-info")
	transferInfoBytes, err := os.ReadFile(transferInfoLoc)
	if err != nil {
		panic(err)
	}

	transferInfo := TransferInfo{}

	if err := yaml.Unmarshal(transferInfoBytes, &transferInfo); err != nil {
		panic(err)
	}

	params.TransferInfo = transferInfo

	log.Println("[INFO] creating Transfer packages")
	results, err := ProcessWorkOrderRows(params, 5)
	if err != nil {
		panic(err)
	}

	//create an output file
	log.Println("[INFO] creating output report")
	outputFile, err := os.Create("adoc-preprocess.tsv")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	writer := csv.NewWriter(outputFile)
	writer.Comma = '\t'
	writer.Write([]string{"worker_id", "component_id", "result", "error"})
	for _, result := range results {
		writer.Write(result)
	}
	writer.Flush()

	log.Printf("[INFO] adoc-preprocess complete for %s_%s", partnerCode, resourceCode)

}
