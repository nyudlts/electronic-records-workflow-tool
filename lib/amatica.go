package lib

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/nyudlts/go-aspace"
	"gopkg.in/yaml.v2"
)

var (
	numWorkers       int
	params           Params
	infectedFilesPtn = regexp.MustCompile("\nInfected files: 0\n")
)

func PrintXferPackageSize(directories bool) error {
	fmt.Println("ewt amatica size, version", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	if err := getPackageSize(config.XferLoc); err != nil {
		return err
	}

	if directories {
		if err := printDirectoryStats(config.XferLoc); err != nil {
			return err
		}
	}

	return nil
}

func PrepAmatica(nWorkers int) error {

	fmt.Println("ewt amatica prep,", VERSION)

	if err := loadConfig(); err != nil {
		return err
	}

	numWorkers = nWorkers

	params = Params{}
	params.Source = config.SIPLoc
	params.XferLoc = config.XferLoc
	params.PartnerCode = config.PartnerCode
	params.ResourceCode = config.CollectionCode

	//locate the work order
	if err := findWorkOrder(); err != nil {
		return err
	}

	mdDir := filepath.Join(config.SIPLoc, "metadata")
	var err error
	params.WorkOrder, err = parseWorkOrder(mdDir, filepath.Base(workOrderLocation))
	if err != nil {
		return err
	}

	//create the transfer-info struct
	transferInfoLoc := filepath.Join(mdDir, "transfer-info.txt")
	transferInfoBytes, err := os.ReadFile(transferInfoLoc)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(transferInfoBytes, &transferInfo); err != nil {
		return err
	}

	params.TransferInfo = transferInfo

	log.Println("[INFO] creating Transfer packages")
	results, err := processWorkOrderRows()
	if err != nil {
		return err
	}

	//create an output log
	log.Println("[INFO] creating output report")
	outputTSVfilename := fmt.Sprintf("%s-xip-prep.tsv", params.ResourceCode)
	outputFile, err := os.Create(filepath.Join(config.LogLoc, outputTSVfilename))
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

func processWorkOrderRows() ([][]string, error) {

	//chunk the workorder rows
	log.Println("[INFO] chunking work order rows")
	chunks := chunkRows(params.WorkOrder.Rows)

	resultChan := make(chan [][]string)

	for i, chunk := range chunks {
		go processChunk(chunk, resultChan, i+1)
	}

	results := [][]string{}
	for range chunks {
		chunkResult := <-resultChan
		results = append(results, chunkResult...)
	}

	return results, nil

}

func chunkRows(rows []aspace.WorkOrderRow) [][]aspace.WorkOrderRow {

	var divided [][]aspace.WorkOrderRow

	chunkSize := (len(rows) + numWorkers - 1) / numWorkers

	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize

		if end > len(rows) {
			end = len(rows)
		}

		divided = append(divided, rows[i:end])
	}

	log.Printf("[INFO] create %d workorder row chunks", len(divided))
	return divided
}

func processChunk(rows []aspace.WorkOrderRow, resultChan chan [][]string, workerId int) {
	results := [][]string{}
	for _, row := range rows {
		if err := createERPackage(row, workerId); err != nil {
			results = append(results, []string{fmt.Sprintf("%d", workerId), row.GetComponentID(), "ERROR", strings.ReplaceAll(err.Error(), "\n", "")})
			continue
		}
		results = append(results, []string{fmt.Sprintf("%d", workerId), row.GetComponentID(), "SUCCESS"})
	}
	resultChan <- results
}

func createERPackage(row aspace.WorkOrderRow, workerId int) error {
	erID := row.GetComponentID()
	log.Printf("[INFO] WORKER %d processing %s", workerId, erID)
	fmt.Printf("  * WORKER %d processing %s\n", workerId, erID)

	//create the directory in the xfer to amatica location
	log.Printf("[INFO] WORKER %d creating directory in xfer location %s", workerId, erID)
	ERDirName := fmt.Sprintf("%s_%s", params.ResourceCode, erID)
	ERLoc := filepath.Join(params.XferLoc, ERDirName)
	if err := os.Mkdir(ERLoc, 0755); err != nil {
		return err
	}

	//create the metadata directory
	log.Printf("[INFO] WORKER %d creating metadata directory in %s", workerId, erID)
	ERMDDirLoc := filepath.Join(ERLoc, "metadata")
	if err := os.Mkdir(ERMDDirLoc, 0775); err != nil {
		return err
	}

	//copy the transfer-info.txt files
	log.Printf("[INFO] WORKER %d copying transfer-info.txt to metadata directory in %s", workerId, erID)
	mdSourceFile := filepath.Join(params.Source, "metadata", "transfer-info.txt")
	mdTarget := filepath.Join(ERMDDirLoc, "transfer-info.txt")
	_, err := copyFile(mdSourceFile, mdTarget)
	if err != nil {
		return (err)
	}

	//create the workorder for ER
	log.Printf("[INFO] WORKER %d creating workorder in metadata directory in %s", workerId, erID)

	woLocation := filepath.Join(ERMDDirLoc, fmt.Sprintf("%s_%s_aspace_wo.tsv", params.ResourceCode, erID))
	woFile, err := os.Create(woLocation)
	if err != nil {
		return err
	}
	defer woFile.Close()
	csvWriter := csv.NewWriter(woFile)
	csvWriter.Comma = '\t'
	csvWriter.Write(aspace.HEADER_ROW)
	csvWriter.Write(getStringArray(row))
	csvWriter.Flush()

	//create the DC json
	log.Printf("[INFO] WORKER %d creating dc.json in metadata directory in %s", workerId, erID)
	dc := createDC(params.TransferInfo, row)
	dcBytes, err := json.Marshal(dc)
	if err != nil {
		return err
	}
	dcLocation := filepath.Join(ERMDDirLoc, "dc.json")
	if err := os.WriteFile(dcLocation, dcBytes, 0755); err != nil {
		return (err)
	}

	//check for and copy FTK CSV
	ftkCSV := fmt.Sprintf("%s.tsv", erID)
	ftkCSVLocation := filepath.Join(params.Source, "metadata", ftkCSV)
	_, err = os.Stat(ftkCSVLocation)
	if err != nil {
		log.Printf("[INFO] WORKER %d no ftk csv in metadata directory in %s", workerId, erID)
	} else {
		log.Printf("[INFO] WORKER %d copying FTK CSV to metadata directory in %s", workerId, erID)
		ftkCSVTarget := filepath.Join(ERMDDirLoc, fmt.Sprintf("%s-ftk.tsv", erID))
		_, err := copyFile(ftkCSVLocation, ftkCSVTarget)
		if err != nil {
			return (err)
		}
	}

	//check for and copy Clamscan logs
	clamscanLog := fmt.Sprintf("%s_clamscan.log", erID)
	clamscanLogLocation := filepath.Join(params.Source, "metadata", clamscanLog)
	_, err = os.Stat(clamscanLogLocation)
	if err != nil {
		log.Printf("[INFO] WORKER %d no clamscan log in metadata directory in %s", workerId, erID)
	} else {
		if !checkClamscanLog(clamscanLogLocation) {
			return fmt.Errorf("clamscan.txt contained infected files")
		}
		log.Printf("[INFO] WORKER %d copying clamscan log to metadata directory in %s", workerId, erID)
		clamscanLogTarget := filepath.Join(ERMDDirLoc, clamscanLog)
		_, err := copyFile(clamscanLogLocation, clamscanLogTarget)
		if err != nil {
			return err
		}
	}

	log.Printf("[INFO] WORKER %d moving payload %s to xfer dir", workerId, erID)
	// move the payload directory to to er directory
	payloadSource := filepath.Join(config.SIPLoc, erID)
	payloadTarget := filepath.Join("xfer", ERDirName, erID)
	fmt.Printf("    source: %s\n    target: %s\n", payloadSource, payloadTarget)

	if err := os.Rename(payloadSource, payloadTarget); err != nil {
		return err
	}

	log.Printf("[INFO] WORKER %d %s complete", workerId, erID)
	fmt.Printf("  * WORKER %d completed %s\n", workerId, erID)
	return nil

}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func getStringArray(row aspace.WorkOrderRow) []string {
	return []string{row.GetResourceID(), row.GetRefID(), row.GetURI(), row.GetContainerIndicator1(), row.GetContainerIndicator2(), row.GetContainerIndicator3(), row.GetTitle(), row.GetComponentID()}
}

func createDC(transferInfo TransferInfo, row aspace.WorkOrderRow) DC {
	dc := DC{}
	dc.IsPartOf = fmt.Sprintf("AIC#%s: %s", transferInfo.ResourceID, transferInfo.ResourceTitle)
	dc.Title = row.GetTitle()
	return dc
}

func checkClamscanLog(logPath string) bool {
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		panic(err)
	}

	if infectedFilesPtn.Match(logBytes) {
		return true
	}

	return false
}
