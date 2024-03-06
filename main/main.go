package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	amtp "github.com/nyudlts/amatica-transfer-prep"
	"github.com/nyudlts/go-aspace"
	"gopkg.in/yaml.v2"
)

const version = "0.1.0"

var (
	partner      string
	resourceCode string
	source       string
	stagingLoc   string
	transferInfo amtp.TransferInfo
)

func init() {
	flag.StringVar(&source, "source", "", "")
	flag.StringVar(&stagingLoc, "staging", "", "")
}

func main() {
	fmt.Printf("Archivematica Transfer Prep v%s\n", version)
	flag.Parse()
	params := amtp.Params{}

	logFile, err := os.Create("amtp.log")
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

	if err := yaml.Unmarshal(transferInfoBytes, &transferInfo); err != nil {
		panic(err)
	}
	params.TransferInfo = transferInfo

	log.Println("INFO creating Transfer packages")
	results, err := amtp.ProcessWorkOrderRows(workOrder, params, 5)
	if err != nil {
		panic(err)
	}

	//create an output file
	log.Println("INFO creating output report")
	outputFile, err := os.Create("output.txt")
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()
	writer := bufio.NewWriter(outputFile)

	for _, result := range results {
		writer.WriteString(result)
		writer.Flush()
	}

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
