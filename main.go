package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const woHeader = "Resource ID	Ref ID	URI	Container Indicator 1	Container Indicator 2	Container Indicator 3	Title	Component ID\n"

var (
	partner      string
	resourceCode string
	source       string
	stagingLoc   string
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
	workorderName, err := getWorkOrder(mdDir)
	if err != nil {
		panic(err)
	}

	partner, resourceCode = getPartnerAndResource(workorderName)

	//get work order components
	workOrderLoc := filepath.Join(mdDir, *workorderName)
	components, err := getWorkOrderComponents(workOrderLoc)
	if err != nil {
		panic(err)
	}

	//process the components
	for _, component := range components {
		err := createERPackage(component)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func createERPackage(component WorkOrderComponent) error {
	erID, err := component.GetERID()
	if err != nil {
		return err
	}

	//create the staging directory
	ERDirName := fmt.Sprintf("%s_%s_electronic-records-%d", partner, resourceCode, erID)
	ERLoc := filepath.Join(stagingLoc, ERDirName)
	if err := os.Mkdir(ERLoc, 0755); err != nil {
		return err
	}

	//create the metadata directory
	ERMDDirLoc := filepath.Join(ERLoc, "metadata")
	fmt.Println(ERMDDirLoc)
	if err := os.Mkdir(ERMDDirLoc, 0755); err != nil {
		return err
	}

	//copy the metadata files
	for _, mdFile := range []string{"dc.json", "transfer-info.txt"} {
		mdSourceFile := filepath.Join(source, "metadata", mdFile)
		mdTarget := filepath.Join(ERMDDirLoc, mdFile)
		_, err = copyFile(mdSourceFile, mdTarget)
		if err != nil {
			return (err)
		}
	}

	//create the workorder
	woOutput := fmt.Sprintf("%s%s", woHeader, component)
	if err != nil {
		return err
	}
	woLocation := filepath.Join(ERMDDirLoc, fmt.Sprintf("%s_%s_electronic_records_%d_aspace_wo.tsv", partner, resourceCode, erID))
	if err := os.WriteFile(woLocation, []byte(woOutput), 0755); err != nil {
		return err
	}

	//create the ER Directory
	dataDir := filepath.Join(ERLoc, component.ContainerIndicator2)
	if err := os.Mkdir(dataDir, 0755); err != nil {
		return err
	}

	//copy files from source to target
	dataSource := filepath.Join(source, component.ContainerIndicator2)
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

func getWorkOrder(path string) (*string, error) {
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

func getWorkOrderComponents(workOrderFile string) ([]WorkOrderComponent, error) {
	workOrderComponents := []WorkOrderComponent{}
	workOrder, err := os.Open(workOrderFile)
	if err != nil {
		return workOrderComponents, err
	}
	defer workOrder.Close()

	scanner := bufio.NewScanner(workOrder)
	scanner.Scan() // skip the header
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), "\t")
		workOrderComponents = append(workOrderComponents, WorkOrderComponent{line[0], line[1], line[2], line[3], line[4], line[5], line[6], line[7]})
	}
	return workOrderComponents, nil
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

func CreateDC(transferInfo TransferInfo, workOrderComponent WorkOrderComponent) DC {
	dc := DC{}
	dc.IsPartOf = fmt.Sprintf("AIC#%s: %s", transferInfo.ResourceID, transferInfo.ResourceTitle)
	dc.Title = workOrderComponent.Title
	return dc
}
