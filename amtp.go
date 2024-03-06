package amtp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyudlts/go-aspace"
	cp "github.com/otiai10/copy"
)

var options = cp.Options{}
var params Params

type Params struct {
	Partner      string
	ResourceCode string
	Source       string
	StagingLoc   string
	TransferInfo TransferInfo
}
type DC struct {
	Title    string `json:"title"`
	IsPartOf string `json:"is_part_of"`
}

type TransferInfo struct {
	ContactName              string `yaml:"Contact-Name"`
	ContactPhone             string `yaml:"Contact-Phone"`
	ContactEmail             string `yaml:"Contact-Email"`
	InternalSenderIdentifier string `yaml:"Internal-Sender-Identifier"`
	OrganizationAdress       string `yaml:"OrganizationAddress"`
	SourceOrganization       string `yaml:"SourceOrganization"`
	ArchivesSpaceResourceURL string `yaml:"nyu-dl-archivesspace-resource-url"`
	ResourceID               string `yaml:"nyu-dl-resource-id"`
	ResourceTitle            string `yaml:"nyu-dl-resource-title"`
	ContrentType             string `yaml:"nyu-dl-content-type"`
	ContentClassification    string `yaml:"nyu-dl-content-classification"`
	ProjectName              string `yaml:"nyu-dl-project-name"`
	RStarCollectionID        string `yaml:"nyu-dl-rstar-collection-id"`
}

func ProcessWorkOrderRows(workOrder aspace.WorkOrder, p Params, numWorkers int) ([]string, error) {
	params = p
	options.PreserveTimes = true
	options.NumOfWorkers = int64(numWorkers)

	//chunk the workorder rows
	log.Println("INFO chunking work order rows")
	chunks := chunkRows(workOrder.Rows, numWorkers)

	resultChan := make(chan []string)

	for i, chunk := range chunks {
		go processChunk(chunk, resultChan, i+1)
	}

	results := []string{}
	for range chunks {
		chunkResult := <-resultChan
		results = append(results, chunkResult...)
	}

	return results, nil

}

func chunkRows(rows []aspace.WorkOrderRow, numWorkers int) [][]aspace.WorkOrderRow {

	var divided [][]aspace.WorkOrderRow

	chunkSize := (len(rows) + numWorkers - 1) / numWorkers

	for i := 0; i < len(rows); i += chunkSize {
		end := i + chunkSize

		if end > len(rows) {
			end = len(rows)
		}

		divided = append(divided, rows[i:end])
	}

	log.Printf("INFO create %d workorder row chunks", len(divided))
	return divided
}

func processChunk(rows []aspace.WorkOrderRow, resultChan chan []string, workerId int) {
	results := []string{}
	for _, row := range rows {
		if err := createERPackage(row, workerId); err != nil {
			results = append(results, fmt.Sprintf("%s\t%s\n", row.GetComponentID(), err))
			continue
		}
		results = append(results, fmt.Sprintf("%s\t%s\n", row.GetComponentID(), "SUCCESS"))
	}
	resultChan <- results
}

func createERPackage(row aspace.WorkOrderRow, workerId int) error {
	erID := row.GetComponentID()
	log.Printf("INFO WORKER %d processing %s", workerId, erID)
	fmt.Printf("* WORKER %d processing %s\n", workerId, erID)

	//create the staging directory
	log.Printf("INFO WORKER %d creating directory in staging location %s", workerId, erID)
	ERDirName := fmt.Sprintf("%s_%s_%s", params.Partner, params.ResourceCode, erID)
	ERLoc := filepath.Join(params.StagingLoc, ERDirName)
	if err := os.Mkdir(ERLoc, 0755); err != nil {
		return err
	}

	//create the metadata directory
	log.Printf("INFO WORKER %d creating metadata directory", workerId)
	ERMDDirLoc := filepath.Join(ERLoc, "metadata")
	if err := os.Mkdir(ERMDDirLoc, 0755); err != nil {
		return err
	}

	//copy the transfer-info.txt files
	log.Printf("INFO WORKER %d copying transfer-info.txt", workerId)
	mdSourceFile := filepath.Join(params.Source, "metadata", "transfer-info.txt")
	mdTarget := filepath.Join(ERMDDirLoc, "transfer-info.txt")
	_, err := copyFile(mdSourceFile, mdTarget)
	if err != nil {
		return (err)
	}

	//create the workorder
	log.Printf("INFO WORKER %d creating workorder", workerId)
	woOutput := fmt.Sprintf("%s\n%s\n", strings.Join(aspace.HEADER_ROW, "\t"), row)

	woLocation := filepath.Join(ERMDDirLoc, fmt.Sprintf("%s_%s_%s_aspace_wo.tsv", params.Partner, params.ResourceCode, erID))
	if err := os.WriteFile(woLocation, []byte(woOutput), 0755); err != nil {
		return err
	}

	//create the DC json
	log.Printf("INFO WORKER %d creating dc.json", workerId)
	dc := CreateDC(params.TransferInfo, row)
	dcBytes, err := json.Marshal(dc)
	if err != nil {
		return err
	}
	dcLocation := filepath.Join(ERMDDirLoc, "dc.json")
	if err := os.WriteFile(dcLocation, dcBytes, 0755); err != nil {
		return (err)
	}

	//create the ER Directory
	log.Printf("INFO WORKER %d creating data directory %s", workerId, erID)
	dataDir := filepath.Join(ERLoc, erID)
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return err
	}

	//copy files from source to target
	payloadSource := filepath.Join(params.Source, erID)
	payloadTarget := (filepath.Join(dataDir))
	log.Printf("INFO WORKER %d copying %s to payload", workerId, erID)
	if err := cp.Copy(payloadSource, payloadTarget, options); err != nil {
		return err
	}

	//complete
	log.Printf("INFO WORKER %d %s complete", workerId, erID)
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

func CreateDC(transferInfo TransferInfo, row aspace.WorkOrderRow) DC {
	dc := DC{}
	dc.IsPartOf = fmt.Sprintf("AIC#%s: %s", transferInfo.ResourceID, transferInfo.ResourceTitle)
	dc.Title = row.GetTitle()
	return dc
}
