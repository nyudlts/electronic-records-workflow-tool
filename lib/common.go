package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyudlts/go-aspace"
	"gopkg.in/yaml.v2"
)

var (
	config            = Config{}
	workOrderLocation string
)

const VERSION = "v1.1.0"

func loadConfig() error {
	//read the adoc-config
	b, err := os.ReadFile("config.yml")
	if err != nil {
		return err
	}

	//unmarshal to config options
	if err := yaml.Unmarshal(b, &config); err != nil {
		return err
	}

	return nil
}

func findWorkOrder() error {
	mdDir := filepath.Join(config.SIPLoc, "metadata")
	var err error
	workOrderFilename, err := getWorkOrderFile(mdDir)
	if err != nil {
		return err
	}
	workOrderLocation = filepath.Join(mdDir, workOrderFilename)
	return nil
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

// model definitions
type Config struct {
	SIPLoc           string `yaml:"sip-location"`
	SourceLoc        string `yaml:"source-location"`
	PartnerCode      string `yaml:"partner-code"`
	CollectionCode   string `yaml:"collection-code"`
	ProjectLoc       string `yaml:"project-location"`
	LogLoc           string `yaml:"log-location"`
	AIPLoc           string `yaml:"aip-location"`
	AMTransferSource string `yaml:"archivematica-transfer-source"`
	XferLoc          string `yaml:"xfer-location"`
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

func (t *TransferInfo) GetResourceID() string {
	split := strings.Split(t.ArchivesSpaceResourceURL, "/")
	return split[len(split)-1]
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

type DC struct {
	Title    string `json:"title"`
	IsPartOf string `json:"is_part_of"`
}
