package lib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/nyudlts/go-aspace"
	"gopkg.in/yaml.v2"
)

var (
	config                   = Config{}
	workOrderLocation        string
	aspaceResourceURLPtn     = regexp.MustCompile(`^/repositories/[2|3|6|99]/resources/\d*$`)
	partnerPtn               = regexp.MustCompile(`^[tamwag|fales|nyuarchives|dlts]`)
	contentClassificationPtn = regexp.MustCompile(`[open|closed|restricted]`)
	packageFormatPtn         = regexp.MustCompile(`["1.0.0"|"1.0.1"]`)
	contentTypePtn           = regexp.MustCompile(`electronic_records|electronic_records-do-not-create-DOs`)
	transferTypePtn          = regexp.MustCompile(`[AIP|XIP]`)
	useStatementPtn          = regexp.MustCompile(`electronic-records-reading-room`)
	clamInfectedPtn          = regexp.MustCompile("\nInfected files: 0")
)

const VERSION = "v1.2.0-alpha"

func loadConfig() error {
	if err := loadConfigPath("config.json"); err != nil {
		return err
	}
	return nil
}

func loadConfigPath(configPath string) error {
	//read the ewt-config
	b, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	//unmarshal to config options
	if err := json.Unmarshal(b, &config); err != nil {
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

func getTransferInfo() (TransferInfo, error) {
	transferInfo = TransferInfo{}
	transferInfoLoc := filepath.Join(config.SIPLoc, "metadata", "transfer-info.txt")
	transferInfoBytes, err := os.ReadFile(transferInfoLoc)
	if err != nil {
		return transferInfo, err
	}

	if err := yaml.Unmarshal(transferInfoBytes, &transferInfo); err != nil {
		return transferInfo, err
	}

	return transferInfo, nil
}

// model definitions
type Config struct {
	SIPLoc           string `json:"sip-location"`
	SourceLoc        string `json:"source-location"`
	PartnerCode      string `json:"partner-code"`
	CollectionCode   string `json:"collection-code"`
	ProjectLoc       string `json:"project-location"`
	LogLoc           string `json:"log-location"`
	AIPLoc           string `json:"aip-location"`
	AMTransferSource string `json:"archivematica-transfer-source"`
	XferLoc          string `json:"xfer-location"`
	AIPStoreLoc      string `json:"aipstore-location"`
	WorkLoc          string `json:"work-location"`
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
	Title      string `json:"title"`
	IsPartOf   string `json:"is_part_of"`
	Identifier string `json:"identifier"`
}

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

type LogType int

const (
	AIP_FILE LogType = iota
	AMATICA_CLEAR
	AMATICA_PREP
	AMATICA_TRANSFER
	ASPACE_CHECK
	RSTAR_TRANSFER
	RSTAR_VALIDATE
	RSTAR_PREP_PACKAGES
	RSTAR_PREP_PACKAGE
	SIP_SCAN_CLEAN
	SIP_SCAN_AV
	SIP_SCAN_DETOX
	SIP_SCAN_CHARS
	SOURCE_TRANSFER
	SIP_VALIDATE
)

func GetLog(logType LogType) string {
	var logPath string
	switch logType {
	case AIP_FILE:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "aip-file.txt"))
	case AMATICA_CLEAR:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "amatica-clear.log"))
	case AMATICA_PREP:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "amatica-prep.log"))
	case AMATICA_TRANSFER:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "amatica-transfer.log"))
	case ASPACE_CHECK:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "aspace-check.tsv"))
	case RSTAR_TRANSFER:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "rstar-transfer.txt"))
	case RSTAR_VALIDATE:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "rstar-validate.log"))
	case RSTAR_PREP_PACKAGES:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "rstar-prep-packages.log"))
	case RSTAR_PREP_PACKAGE:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "rstar-prep-single.log"))
	case SIP_SCAN_CLEAN:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "sip-scan-clean.log"))
	case SIP_SCAN_AV:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "sip-scan-av.log"))
	case SIP_SCAN_DETOX:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "sip-scan-detox.log"))
	case SIP_SCAN_CHARS:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "sip-scan-chars.log"))
	case SIP_VALIDATE:
		logPath = filepath.Join(config.LogLoc, fmt.Sprintf("%s-%s", config.CollectionCode, "sip-validate.log"))
	case SOURCE_TRANSFER:
		logPath = filepath.Join(config.LogLoc, "rsync", fmt.Sprintf("%s-%s", config.CollectionCode, "source-transfer-rsync.txt"))
	}
	return logPath
}

func ReadLog(logType LogType) error {
	if err := loadConfig(); err != nil {
		return err
	}

	fmt.Printf("ewt log reader, version %s\n", VERSION)
	log := GetLog(logType)
	logName := filepath.Base(log)
	fmt.Printf("  * reading %s\n\n", logName)

	if err := printLog(log); err != nil {
		return err
	}
	return nil
}

func printLog(logPath string) error {
	logBytes, err := os.ReadFile(logPath)
	if err != nil {
		return err
	}
	fmt.Println(string(logBytes))
	return nil
}
