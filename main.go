package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyudlts/go-aspace"
	"gopkg.in/yaml.v2"
)

var (
	partner      string
	resourceCode string
	source       string
	stagingLoc   string
	transferInfo TransferInfo
)

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

	//create the staging directory
	ERDirName := fmt.Sprintf("%s_%s_%s", partner, resourceCode, erID)
	ERLoc := filepath.Join(stagingLoc, ERDirName)
	if err := os.Mkdir(ERLoc, 0755); err != nil {
		return err
	}

	//create the metadata directory
	ERMDDirLoc := filepath.Join(ERLoc, "metadata")
	if err := os.Mkdir(ERMDDirLoc, 0755); err != nil {
		return err
	}

	//copy the transfer-info.txt files
	mdSourceFile := filepath.Join(source, "metadata", "transfer-info.txt")
	mdTarget := filepath.Join(ERMDDirLoc, "transfer-info.txt")
	_, err := copyFile(mdSourceFile, mdTarget)
	if err != nil {
		return (err)
	}

	//create the workorder
	woOutput := fmt.Sprintf("%s\n%s\n", strings.Join(aspace.HEADER_ROW, "/t"), row)
	if err != nil {
		return err
	}

	woLocation := filepath.Join(ERMDDirLoc, fmt.Sprintf("%s_%s_%s_aspace_wo.tsv", partner, resourceCode, erID))
	if err := os.WriteFile(woLocation, []byte(woOutput), 0755); err != nil {
		return err
	}

	//create the DC json
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
	dataDir := filepath.Join(ERLoc, erID)
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return err
	}

	//copy files from source to target
	dataSource := filepath.Join(source, erID)
	sourceFiles, err := os.ReadDir(dataSource)
	if err != nil {
		return err
	}

	for _, sourceFile := range sourceFiles {
		sourceDataFile := filepath.Join(dataSource, sourceFile.Name())
		targetDataFile := filepath.Join(dataDir, sourceFile.Name())
		_, err := copyFile(sourceDataFile, targetDataFile)
		if err != nil {
			return err
		}
	}

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
