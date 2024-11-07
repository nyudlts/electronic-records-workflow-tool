package cmd

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var numWorkers int

func init() {
	stageCmd.Flags().StringVar(&sourceLoc, "source-location", "", "the location of the package to be transferred to r*")
	stageCmd.Flags().StringVar(&stagingLoc, "staging-location", "", "the locatin to copy packages to")
	stageCmd.Flags().IntVar(&numWorkers, "workers", 1, "")
	rootCmd.AddCommand(stageCmd)
}

var stageCmd = &cobra.Command{
	Use:   "stage",
	Short: "pre-process submission package",
	Run: func(cmd *cobra.Command, args []string) {
		stage()
	},
}

type Params struct {
	PartnerCode  string
	ResourceCode string
	Source       string
	Staging      string
	TransferInfo TransferInfo
	WorkOrder    aspace.WorkOrder
}

func stage() {
	fmt.Printf("adoc-process %s\n", version)
	flag.Parse()
	params := Params{}

	logFileName := "adoc-stage.log"
	logFile, err := os.Create(logFileName) //this should have the call number of the collection e.g. fales_mss318.log
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

	log.Println("[INFO] checking the staging directory exists")
	//check that source exists and is a Directory
	if err := isDirectory(stagingLoc); err != nil {
		panic(err)
	}

	params.Staging = stagingLoc

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
	results, err := ProcessWorkOrderRows(params, numWorkers)
	if err != nil {
		panic(err)
	}

	//create an output file
	log.Println("[INFO] creating output report")
	outputFile, err := os.Create(fmt.Sprintf("%s_%s-adoc-stage.tsv", params.PartnerCode, params.ResourceCode))
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

	log.Printf("[INFO] adoc-stage complete for %s_%s", params.PartnerCode, params.ResourceCode)

}
