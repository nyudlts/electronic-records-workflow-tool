package cmd

import (
	"encoding/csv"
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
	stageCmd.Flags().IntVar(&numWorkers, "workers", 1, "number of worker threads to process SIPs")
	sipCmd.AddCommand(stageCmd)
}

var stageCmd = &cobra.Command{
	Use:   "prep",
	Short: "generate SIPs to transfer to archivematica",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("ADOC SIP Prep")

		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		if err := stage(); err != nil {
			panic(err)
		}
	},
}

type Params struct {
	PartnerCode  string
	ResourceCode string
	Source       string
	Staging      string
	TransferInfo TransferInfo
	WorkOrder    aspace.WorkOrder
	XferLoc      string
}

func stage() error {
	fmt.Printf("adoc stage %s\n", version)
	params := Params{}
	msgs := []string{}

	msgs = append(msgs, "[INFO] checking staging directory exists")
	//check that source exists and is a Directory
	if err := isDirectory(adocConfig.StagingLoc); err != nil {
		return err
	}
	params.Source = adocConfig.StagingLoc

	msgs = append(msgs, "[INFO] checking the xfer directory exists")
	xferLoc := filepath.Join(adocConfig.ProjectLoc, "xfer")
	//check that source exists and is a Directory
	if err := isDirectory(xferLoc); err != nil {
		return err
	}
	params.XferLoc = xferLoc

	msgs = append(msgs, "[INFO] checking metadata directory exists")
	//check that metadata directory exists and is a directory
	mdDir := filepath.Join(adocConfig.StagingLoc, "metadata")
	if err := isDirectory(mdDir); err != nil {
		return err
	}

	msgs = append(msgs, "[INFO] checking work order exists")
	//find a work order
	workorderName, err := getWorkOrderFile(mdDir)
	if err != nil {
		return err
	}

	//getting partner and resource code
	msgs = append(msgs, "[INFO] getting partner and resource code")
	params.PartnerCode = adocConfig.PartnerCode
	params.ResourceCode = adocConfig.CollectionCode

	//create the logfile
	logFileName := fmt.Sprintf("%s-adoc-prep-sip.log", params.ResourceCode)
	logFile, err := os.Create(filepath.Join(adocConfig.LogLoc, logFileName))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//write the log messages
	for _, msg := range msgs {
		log.Println(msg)
	}

	//load the work order
	log.Println("[INFO] parsing work order")

	params.WorkOrder, err = parseWorkOrder(mdDir, workorderName)
	if err != nil {
		return err
	}

	//find the transfer info file
	log.Println("[INFO] checking transfer info exists")
	transferInfoLoc := filepath.Join(mdDir, "transfer-info.txt")
	if _, err = os.Stat(transferInfoLoc); err != nil {
		return err
	}

	//create the transfer-info struct
	log.Println("[INFO] parsing transfer-info.txt")
	transferInfoBytes, err := os.ReadFile(transferInfoLoc)
	if err != nil {
		return err
	}

	transferInfo := TransferInfo{}

	if err := yaml.Unmarshal(transferInfoBytes, &transferInfo); err != nil {
		return err
	}

	params.TransferInfo = transferInfo

	log.Println("[INFO] creating Transfer packages")
	results, err := ProcessWorkOrderRows(params, numWorkers)
	if err != nil {
		return err
	}

	//create an output file
	log.Println("[INFO] creating output report")
	outputTSVfilename := fmt.Sprintf("%s-adoc-prep-sip.tsv", params.ResourceCode)
	outputFile, err := os.Create(filepath.Join(adocConfig.LogLoc, outputTSVfilename))
	if err != nil {
		return err
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

	return nil

}
