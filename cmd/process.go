package cmd

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const version = "0.1.1"

var (
	partner      string
	resourceCode string
	source       string
	stagingLoc   string
	transferInfo TransferInfo
)

func init() {
	processCmd.PersistentFlags().StringVar(&source, "source", "", "")
	processCmd.PersistentFlags().StringVar(&stagingLoc, "staging", "", "")
	rootCmd.AddCommand(processCmd)
}

var processCmd = &cobra.Command{
	Use: "process",
	Run: func(cmd *cobra.Command, args []string) {
		process()
	},
}

func process() {
	fmt.Printf("Archivematica Transfer Prep v%s\n", version)
	flag.Parse()
	params := Params{}

	logFile, err := os.Create("adoc-preprocess.log")
	if err != nil {
		panic(err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)

	log.Println("INFO checking source directory")
	//check that source exists and is a Directory
	if err := isDirectory(source); err != nil {
		panic(err)
	}
	params.Source = source

	//check that staging location exists and is a Directory
	if err := isDirectory(stagingLoc); err != nil {
		panic(err)
	}
	params.StagingLoc = stagingLoc

	log.Println("INFO checking metadata directory")
	//check that metadata directory exists and is a directory
	mdDir := filepath.Join(source, "metadata")
	if err := isDirectory(mdDir); err != nil {
		panic(err)
	}

	log.Println("INFO locating work order")

	//find a work order
	workorderName, err := getWorkOrderFile(mdDir)
	if err != nil {
		panic(err)
	}

	//getting partner and resource code
	log.Println("INFO getting partner and resource code")
	partner, resourceCode = getPartnerAndResource(workorderName)
	params.Partner = partner
	params.ResourceCode = resourceCode

	//load the work order
	log.Println("INFO loading work order")
	workOrderLoc := filepath.Join(mdDir, *workorderName)
	wof, err := os.Open(workOrderLoc)
	if err != nil {
		panic(err)
	}
	defer wof.Close()
	var workOrder aspace.WorkOrder
	if err := workOrder.Load(wof); err != nil {
		panic(err)
	}

	//create the transfer-info struct
	log.Println("INFO creating transfer-info struct")
	transferInfoLoc := filepath.Join(mdDir, "transfer-info.txt")
	transferInfoBytes, err := os.ReadFile(transferInfoLoc)
	if err != nil {
		panic(err)
	}

	transferInfo := TransferInfo{}

	if err := yaml.Unmarshal(transferInfoBytes, &transferInfo); err != nil {
		panic(err)
	}

	params.TransferInfo = transferInfo

	log.Println("INFO creating Transfer packages")
	results, err := ProcessWorkOrderRows(workOrder, params, 5)
	if err != nil {
		panic(err)
	}

	//create an output file
	log.Println("INFO creating output report")
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

}

func isDirectory(path string) error {
	fi, err := os.Stat(path)
	if err == nil {
		if fi.IsDir() {
			return nil
		} else {
			return fmt.Errorf("%s is not a directory", path)
		}
	} else {
		return err
	}
}

func getWorkOrderFile(path string) (*string, error) {
	mdFiles, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, mdFile := range mdFiles {
		name := mdFile.Name()
		if strings.Contains(name, "aspace_wo.tsv") {
			return &name, nil
		}
	}
	return nil, fmt.Errorf("%s does not contain a work order", path)
}

func getPartnerAndResource(workOrderName *string) (string, string) {
	split := strings.Split(*workOrderName, "_")
	return split[0], split[1]
}

var partnerAndCode = regexp.MustCompile(`^[tamwag|fales|nyuarchives//].*`)

func validateTransferInfo(ti *TransferInfo) error {
	//ensure rstar uuid is present
	if _, err := uuid.Parse(transferInfo.RStarCollectionID); err != nil {
		return err
	}

	//ensure that the partner and codes are valid

	return nil
}
