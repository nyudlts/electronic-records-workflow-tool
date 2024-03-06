package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyudlts/go-aspace"
	"gopkg.in/yaml.v2"

	cp "github.com/otiai10/copy"
)

var (
	partner      string
	resourceCode string
	source       string
	stagingLoc   string
	transferInfo TransferInfo
)

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

func init() {
	flag.StringVar(&source, "source", "", "")
	flag.StringVar(&stagingLoc, "staging", "", "")
}

func main() {
	flag.Parse()

	//check that source exists and is a Directory
	if err := isDirectory(source); err != nil {
		panic(err)
	}

	//check that staging location exists and is a Directory
	if err := isDirectory(stagingLoc); err != nil {
		panic(err)
	}

	//check that metadata directory exists and is a directory
	mdDir := filepath.Join(source, "metadata")
	if err := isDirectory(mdDir); err != nil {
		panic(err)
	}

	//find a work order
	workorderName, err := getWorkOrderFile(mdDir)
	if err != nil {
		panic(err)
	}

	partner, resourceCode = getPartnerAndResource(workorderName)
	workOrderLoc := filepath.Join(mdDir, *workorderName)
	wof, err := os.Open(workOrderLoc)
	if err != nil {
		panic(err)
	}
	defer wof.Close()
	var wo aspace.WorkOrder
	if err := wo.Load(wof); err != nil {
		panic(err)
	}

	//create the transfer-info struct
	transferInfoLoc := filepath.Join(mdDir, "transfer-info.txt")
	transferInfoBytes, err := os.ReadFile(transferInfoLoc)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(transferInfoBytes, &transferInfo); err != nil {
		panic(err)
	}

	for _, row := range wo.Rows {
		if err := createERPackage(row); err != nil {
			panic(err)
		}
	}

}

func createERPackage(row aspace.WorkOrderRow) error {
	erID := row.GetComponentID()
	log.Println("INFO\tprocessing", erID)

	//create the staging directory
	ERDirName := fmt.Sprintf("%s_%s_%s", partner, resourceCode, erID)
	ERLoc := filepath.Join(stagingLoc, ERDirName)
	if err := os.Mkdir(ERLoc, 0755); err != nil {
		return err
	}

	//create the metadata directory
	log.Println("INFO\tcreating metadata directory")
	ERMDDirLoc := filepath.Join(ERLoc, "metadata")
	if err := os.Mkdir(ERMDDirLoc, 0755); err != nil {
		return err
	}

	//copy the transfer-info.txt files
	log.Println("INFO\tcopying transfer-info.txt")
	mdSourceFile := filepath.Join(source, "metadata", "transfer-info.txt")
	mdTarget := filepath.Join(ERMDDirLoc, "transfer-info.txt")
	_, err := copyFile(mdSourceFile, mdTarget)
	if err != nil {
		return (err)
	}

	//create the workorder
	log.Println("INFO\tcreating workorder")
	woOutput := fmt.Sprintf("%s\n%s\n", strings.Join(aspace.HEADER_ROW, "\t"), row)

	woLocation := filepath.Join(ERMDDirLoc, fmt.Sprintf("%s_%s_%s_aspace_wo.tsv", partner, resourceCode, erID))
	if err := os.WriteFile(woLocation, []byte(woOutput), 0755); err != nil {
		return err
	}

	//create the DC json
	log.Println("INFO\tcreating dc.json")
	dc := CreateDC(transferInfo, row)
	dcBytes, err := json.Marshal(dc)
	if err != nil {
		return err
	}
	dcLocation := filepath.Join(ERMDDirLoc, "dc.json")
	if err := os.WriteFile(dcLocation, dcBytes, 0755); err != nil {
		return (err)
	}

	//create the ER Directory
	log.Println("INFO\tcreating data directory ", erID)
	dataDir := filepath.Join(ERLoc, erID)
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return err
	}

	//copy files from source to target
	payloadSource := filepath.Join(source, erID)
	payloadTarget := (filepath.Join(dataDir))
	log.Printf("INFO\tcopying %s to payload", erID)
	if err := cp.Copy(payloadSource, payloadTarget); err != nil {
		return err
	}

	log.Printf("INFO\t%s complete", erID)
	return nil
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
