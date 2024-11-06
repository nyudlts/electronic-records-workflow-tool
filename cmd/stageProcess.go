package cmd

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

	"github.com/google/uuid"
	"github.com/nyudlts/go-aspace"
)

var (
	params           Params
	infectedFilesPtn = regexp.MustCompile("\nInfected files: 0\n")
)

type Params struct {
	PartnerCode  string
	ResourceCode string
	Source       string
	TransferInfo TransferInfo
	WorkOrder    aspace.WorkOrder
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
	OrganizationAddress      string `yaml:"Organization-Address"`
	SourceOrganization       string `yaml:"Source-Organization"`
	ArchivesSpaceResourceURL string `yaml:"nyu-dl-archivesspace-resource-url"`
	ResourceID               string `yaml:"nyu-dl-resource-id"`
	ResourceTitle            string `yaml:"nyu-dl-resource-title"`
	ContentType              string `yaml:"nyu-dl-content-type"`
	ContentClassification    string `yaml:"nyu-dl-content-classification"`
	ProjectName              string `yaml:"nyu-dl-project-name"`
	RStarCollectionID        string `yaml:"nyu-dl-rstar-collection-id"`
	PackageFormat            string `yaml:"nyu-dl-package-format"`
	UseStatement             string `yaml:"nyu-dl-use-statement"`
	TransferType             string `yaml:"nyu-dl-transfer-type"`
}

func ProcessWorkOrderRows(p Params, numWorkers int) ([][]string, error) {
	params = p

	//chunk the workorder rows
	log.Println("[INFO] chunking work order rows")
	chunks := chunkRows(p.WorkOrder.Rows, numWorkers)

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
	fmt.Printf("* WORKER %d processing %s\n", workerId, erID)

	//create the staging directory
	log.Printf("[INFO] WORKER %d creating directory in staging location %s", workerId, erID)
	ERDirName := fmt.Sprintf("%s_%s_%s", params.PartnerCode, params.ResourceCode, erID)
	ERLoc := filepath.Join(params.Source, ERDirName)
	if err := os.Mkdir(ERLoc, 0755); err != nil {
		return err
	}

	//create the metadata directory
	log.Printf("[INFO] WORKER %d creating metadata directory in %s", workerId, erID)
	ERMDDirLoc := filepath.Join(ERLoc, "metadata")
	if err := os.Mkdir(ERMDDirLoc, 0755); err != nil {
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

	woLocation := filepath.Join(ERMDDirLoc, fmt.Sprintf("%s_%s_%s_aspace_wo.tsv", params.PartnerCode, params.ResourceCode, erID))
	woFile, err := os.Create(woLocation)
	if err != nil {
		return err
	}
	defer woFile.Close()
	csvWriter := csv.NewWriter(woFile)
	csvWriter.Comma = '\t'
	csvWriter.Write(aspace.HEADER_ROW)
	csvWriter.Write(GetStringArray(row))
	csvWriter.Flush()

	//create the DC json
	log.Printf("[INFO] WORKER %d creating dc.json in metadata directory in %s", workerId, erID)
	dc := CreateDC(params.TransferInfo, row)
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

	// move the payload directory to to er directory
	payloadSource := filepath.Join(params.Source, erID)
	payloadTarget := filepath.Join(ERMDDirLoc, erID)
	if err := os.Rename(payloadSource, payloadTarget); err != nil {
		return err
	}

	/*
		//create the ER Directory

		log.Printf("[INFO] WORKER %d creating data directory %s", workerId, erID)
		dataDir := filepath.Join(ERLoc, erID)
		if err := os.Mkdir(dataDir, 0755); err != nil {
			return err
		}

		//copy files from source to target
		payloadSource := filepath.Join(params.Source, erID)
		payloadTarget := (filepath.Join(dataDir))
		log.Printf("[INFO] WORKER %d copying %s to payload", workerId, erID)

		if err := cp.Copy(payloadSource, payloadTarget, options); err != nil {
			return err
		}
	*/

	//complete
	log.Printf("[INFO] WORKER %d %s complete", workerId, erID)
	fmt.Printf("* WORKER %d completed %s\n", workerId, erID)
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

func GetStringArray(row aspace.WorkOrderRow) []string {
	return []string{row.GetResourceID(), row.GetRefID(), row.GetURI(), row.GetContainerIndicator1(), row.GetContainerIndicator2(), row.GetContainerIndicator3(), row.GetTitle(), row.GetComponentID()}
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

func getWorkOrderFile(path string) (string, error) {
	mdFiles, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}

	for _, mdFile := range mdFiles {
		name := mdFile.Name()
		if strings.Contains(name, "_aspace_wo.tsv") {
			return name, nil
		}
	}
	return "", fmt.Errorf("%s does not contain a work order", path)
}

func getPartnerAndResource(workOrderName string) (string, string) {
	split := strings.Split(workOrderName, "_")
	return split[0], strings.Join(split[1:len(split)-2], "_")
}

// regexp definitions for validation
var (
	aspaceResourceURLPtn     = regexp.MustCompile(`^/repositories/[2|3|6]/resources/\d*$`)
	partnerPtn               = regexp.MustCompile(`^[tamwag|fales|nyuarchives]`)
	contentClassificationPtn = regexp.MustCompile(`[open|closed|restricted]`)
	packageFormatPtn         = regexp.MustCompile(`["1.0.0"|"1.0.1"]`)
	contentTypePtn           = regexp.MustCompile(`electronic_records|electronic_records-do-not-create-DOs`)
	transferTypePtn          = regexp.MustCompile(`[AIP|SIP|DIP]`)
	useStatementPtn          = regexp.MustCompile(`electronic-records-reading-room`)
)

func (ti TransferInfo) Validate() error {
	//ensure contact-name is not blank
	if ti.ContactName == "" {
		return fmt.Errorf("field `Contact-Name` is blank in transfer-info.txt")
	}

	//ensure contact-email is not blank
	if ti.ContactEmail == "" {
		return fmt.Errorf("`Contact-Email` is blank in transfer-info.txt")
	}

	//ensure contact-phone is not blank
	if ti.ContactPhone == "" {
		return fmt.Errorf("`Contact-Phone` is blank in transfer-info.txt")
	}

	//ensure that Internal Sender Identifier is valid
	split := strings.Split(ti.InternalSenderIdentifier, "/")
	if len(split) != 2 {
		return fmt.Errorf("`Internal-Sender-Identifier` is malformed in transfer-info.txt, must contains a single `/`")
	}

	if !partnerPtn.MatchString(split[0]) {
		return fmt.Errorf("`Internal-Sender-Identifier` is malformed in transfer-info.txt, partner code must be one of: `fales`, `tamwag`, or `nyuarchive`")
	}

	//Ensure Source Organization is not blank
	if ti.OrganizationAddress == "" {
		return fmt.Errorf("`Organization-Address` is blank in transfer-info.txt")
	}

	//Ensure Source Organization is not blank
	if ti.SourceOrganization == "" {
		return fmt.Errorf("`Source-Organization` is blank in transfer-info.txt")
	}

	//Ensure there is A ArchivesSpace Resource URL is present and valid
	if !aspaceResourceURLPtn.MatchString(ti.ArchivesSpaceResourceURL) {
		return fmt.Errorf("`nyu-dl-archivesspace-resource-url` malformed in transfer-info.txt, must be in the form `/repositories/X/resources/Y`")
	}

	//Ensure Resource-ID is not blank
	if ti.ResourceID == "" {
		return fmt.Errorf("`nyu-dl-resource-id` is blank in transfer-info.txt")
	}

	//Ensure Resource-Title is not blank
	if ti.ResourceTitle == "" {
		return fmt.Errorf("`nyu-dl-resource-title` is blank in transfer-info.txt")
	}

	//ensure the Content-Type is valid
	if !contentTypePtn.MatchString(ti.ContentType) {
		return fmt.Errorf("`nyu-dl-content-type` must have a value of `electronic_records`, or `electronic_records-do-not-create-DOs`, values was %s", ti.ContentType)
	}

	//ensure the Content-Classification is valid
	if !contentClassificationPtn.MatchString(ti.ContentClassification) {
		return fmt.Errorf("`nyu-dl-content-classification` must have a value of `open`, `closed`, or `restricted`")
	}

	//ensure that the project name is valid
	split = strings.Split(ti.ProjectName, "/")
	if len(split) != 2 {
		return fmt.Errorf("`nyu-dl-project-name` is malformed in transfer-info.txt, must contains a single `/`")
	}

	if !partnerPtn.MatchString(split[0]) {
		return fmt.Errorf("`nyu-dl-project-name` is malformed in transfer-info.txt, partner code must be one of: `fales`, `tamwag`, or `nyuarchive`")
	}

	//ensure rstar uuid is present and valid
	if _, err := uuid.Parse(ti.RStarCollectionID); err != nil {
		return err
	}

	//ensure the package-format is valid
	if !packageFormatPtn.MatchString(ti.PackageFormat) {
		return fmt.Errorf("`nyu-dl-package-format` is malformed in transfer-info.txt, partner code must be one of: `1.0.0`, or 	`1.0.1`")
	}

	//ensure the use-statement is valid
	if !useStatementPtn.MatchString(ti.UseStatement) {
		return fmt.Errorf("`nyu-dl-use-statement` is malformed in transfer-info.txt, use statement must be `electronic-records-reading-room`")
	}

	//ensure the transfer-type is valid
	if !transferTypePtn.MatchString(ti.TransferType) {
		return fmt.Errorf("`nyu-dl-transfer-type` is malformed in transfer-info.txt, transfer type must be one of: `AIP`, `DIP`, or `SIP`")
	}

	return nil
}

func parseWorkOrder(mdDir string, workorderName string) (aspace.WorkOrder, error) {
	workOrderLoc := filepath.Join(mdDir, workorderName)

	wof, err := os.Open(workOrderLoc)
	if err != nil {
		panic(err)
	}
	defer wof.Close()
	var workOrder aspace.WorkOrder
	if err := workOrder.Load(wof); err != nil {
		return workOrder, err
	}
	return workOrder, nil
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
