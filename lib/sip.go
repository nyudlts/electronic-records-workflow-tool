package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/nyudlts/go-aspace"
	"gopkg.in/yaml.v2"
)

var okPattern = regexp.MustCompile(`OK$`)

func PrintSIPPackageSize(directories bool) error {
	fmt.Println("ewt sip size, version", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	if err := getPackageSize(config.SIPLoc); err != nil {
		return err
	}

	if directories {
		if err := printDirectoryStats(config.SIPLoc); err != nil {
			return err
		}
	}

	return nil
}

func CleanSip() error {
	fmt.Println("ewt sip clean, version", VERSION)

	//load the project configuration
	if err := loadConfig(); err != nil {
		return err
	}

	//create a logger
	logFile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s-sip-scan-clean.log", config.CollectionCode)))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	var removeList = []string{".DS_Store", "Thumbs.db", "Desktop.ini", "Icon\r"}

	deleteCount := 0
	if err := filepath.Walk(config.SIPLoc, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			if contains(removeList, info.Name()) {
				if err := os.Remove(path); err != nil {
					return err
				}
				fmt.Printf("  * deleted %q\n", path)
				log.Printf("[INFO] deleted %q\n", path)
				deleteCount++
			}
		}
		return nil
	}); err != nil {
		return err
	}
	fmt.Printf("  * %d files deleted\n", deleteCount)
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func GenerateTransferInfo(profile string) error {
	fmt.Println("ewt sip gen transfer, version", VERSION)
	fmt.Println("  * generating transfer info for profile:", profile)

	if err := loadConfig(); err != nil {
		return err
	}

	profileFilename := strings.ToUpper(profile) + ".txt"
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	parentDirectory, _ := filepath.Split(wd)
	profileFile := filepath.Join(parentDirectory, "templates", profileFilename)

	if _, err := os.Stat(profileFile); err != nil {
		return err
	}

	templateFile := filepath.Join(parentDirectory, "templates", "ewt.txt")
	if _, err := os.Stat(templateFile); err != nil {
		return err
	}

	xferInfo := filepath.Join(config.SIPLoc, "metadata", "transfer-info.txt")

	outFile, err := os.Create(xferInfo)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)

	profileBytes, err := os.ReadFile(profileFile)
	if err != nil {
		return err
	}

	templateBytes, err := os.ReadFile(templateFile)
	if err != nil {
		return err
	}

	writer.Write(profileBytes)
	writer.Write(templateBytes)
	if err := writer.Flush(); err != nil {
		return err
	}

	return nil
}

func ValidateSIP() error {
	fmt.Println("ewt sip validate,", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	var workOrder aspace.WorkOrder

	//create a logger
	logFile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s-sip-validate.log", config.CollectionCode)))
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	log.Printf("[INFO] ewt validate sip %s\n", VERSION)
	fmt.Printf("  * validating SIP at %s\n", config.SIPLoc)
	log.Printf("[INFO] validating SIP transfer package at %s\n", config.SIPLoc)

	hasError := false
	hasWarning := false

	//check that the source directory exists
	fmt.Print("    1. checking that SIP location exists and is a directory: ")
	fileInfo, err := os.Stat(config.SIPLoc)
	if err != nil {
		hasError = true
		log.Printf("[ERROR] %s\n", err.Error())
		fmt.Printf("SIP location %s does not exist, exiting", config.SIPLoc)
		return err
	}

	if !fileInfo.IsDir() {
		log.Printf("[ERROR] %s is not a directory\n", config.SIPLoc)
		fmt.Printf("  * SIP location %s is not a directory, exiting", config.SIPLoc)
		hasError = true
		return fmt.Errorf("%s is not a directory", config.SIPLoc)
	}
	log.Printf("[INFO] %s exists and is a directory", config.SIPLoc)
	fmt.Println(" OK")

	//check that there is a metadata directory
	fmt.Print("    2. checking that SIP directory contains a metadata directory: ")
	mdDirLocation := filepath.Join(config.SIPLoc, "metadata")
	mdDir, err := os.Stat(mdDirLocation)
	if err != nil {
		hasError = true
		fmt.Printf("SIP location %s does not contain a metadata directory", config.SIPLoc)
		log.Printf("[ERROR] %s does not contain a metadata directory\n", config.SIPLoc)
		return (err)
	}

	if !mdDir.IsDir() {
		fmt.Printf("  * %s metadata directory is not a directory\n", mdDirLocation)
		log.Printf("[ERROR] %s is not a directory\n", mdDirLocation)
		hasError = true
		return fmt.Errorf("[ERROR] %s is not a directory\n", mdDirLocation)

	}
	log.Printf("[INFO] %s contains a metadata directory\n", config.SIPLoc)
	fmt.Println("OK")

	//check that a workOrder exists
	fmt.Print("    3. checking that a valid workorder file exists: ")
	workorderName, err := getWorkOrderFile(mdDirLocation)
	if err != nil {
		hasError = true
		fmt.Printf("metadata directory %s does not contain a work order\n", mdDirLocation)
		log.Printf("[ERROR] metadata directory %s does not contain a work order\n", mdDirLocation)
	} else {
		//check that the workorder is valid
		workOrder, err = parseWorkOrder(mdDirLocation, workorderName)
		if err != nil {
			hasError = true
			fmt.Printf("work order %s is not valid: %s\n", mdDirLocation, err.Error())
			log.Printf("[ERROR] work order %s is not valid: %s\n", mdDirLocation, err.Error())
		} else {
			fmt.Println("OK")
			log.Printf("[INFO] check 3. %s contains a valid work order \n", mdDirLocation)
		}
	}

	//get a list of componentIDs from work order
	fmt.Printf("    4. checking workorder %s for duplicate cuids: ", workorderName)
	componentIDs := []string{}
	//get an array of componentIDs
	dupeCount := 0
	for _, row := range workOrder.Rows {
		if woContains(row.GetComponentID(), componentIDs) {
			log.Printf("[ERROR] duplicate componentID, %s, found in workorder\n", row.GetComponentID())
			dupeCount++
		} else {
			componentIDs = append(componentIDs, row.GetComponentID())
		}
	}

	sort.Strings(componentIDs)
	log.Printf("[ERROR] check 4. %s contains %d duplicate cuids \n", workorderName, dupeCount)
	if dupeCount > 0 {
		fmt.Println("ERROR")
	} else {
		fmt.Println("OK")
	}

	//check that all componentIDs in the workorder exist in the SIP
	fmt.Print("    5. checking all ER directories in workorder exist: ")
	missingDirs := 0
	for _, componentID := range componentIDs {
		erLocation := filepath.Join(config.SIPLoc, componentID)
		if _, err := os.Stat(erLocation); err != nil {
			missingDirs++
			log.Printf("[ERROR] componentID, %s is missing in transfered directories\n", componentID)
			//fmt.Printf("  * cuid %s is missing from transferred directories", componentID)
		}
	}
	log.Printf("[ERROR] check 5. %s contains %d missing transfer directories \n", workorderName, missingDirs)

	if missingDirs > 0 {
		hasError = true
		fmt.Println("ERROR")
	} else {
		fmt.Println("OK")
	}

	//check there are no extra directories in source location
	fmt.Print("    6. checking that there are no extra directories or files in SIP directory: ")
	sourceDirs, err := os.ReadDir(config.SIPLoc)
	if err != nil {
		log.Printf("[ERROR] could not read SIP directory %s: %s\n", config.SIPLoc, err.Error())
		fmt.Printf("[ERROR] could not read SIP directory %s: %s\n", config.SIPLoc, err.Error())
		hasError = true
	} else {

		extraDirs := 0
		for _, sourceDir := range sourceDirs {
			if sourceDir.Name() != "metadata" {
				if !woContains(sourceDir.Name(), componentIDs) {
					extraDirs++
					log.Printf("[ERROR] %s is not listed on workorder\n", sourceDir.Name())
				}
			}
		}

		log.Printf("[INFO] check 6. %s contained %d extra objects\n", config.SIPLoc, extraDirs)
		if extraDirs > 0 {
			hasError = true
			fmt.Println("ERROR")
		} else {
			fmt.Println("OK")
		}
	}

	//check that SIP contains a valid transfer-info.txt
	fmt.Print("    7. checking that valid transfer-info.txt exists: ")
	xferInfoLocation := filepath.Join(mdDirLocation, "transfer-info.txt")
	_, err = os.Stat(xferInfoLocation)
	if err != nil {
		fmt.Println("transfer-info.txt does not exist in metadata directory")
		log.Println("[ERROR] transfer-info.txt does not exist in metadata directory")
	} else {
		xferBytes, err := os.ReadFile(xferInfoLocation)
		if err != nil {
			hasError = true
			fmt.Printf("could not read transfer-info.txt: %s\n", xferInfoLocation)
			log.Printf("[ERROR] could not read transfer-info.txt: %s\n", xferInfoLocation)
		} else {
			transferInfo := TransferInfo{}
			if err := yaml.Unmarshal(xferBytes, &transferInfo); err != nil {
				hasError = true
				fmt.Println("could not unmarshal transfer-info.txt")
				log.Println("[ERROR] could not unmarshal transfer-info.txt")
			} else {
				if err := transferInfo.Validate(); err != nil {
					hasError = true
					fmt.Printf("transfer-info.txt is not valid: %s\n", err.Error())
					log.Printf("[ERROR] transfer-info.txt is not valid: %s\n", err.Error())
				} else {
					log.Printf("[INFO] check 7. %s contains a valid transfer-info.txt \n", mdDirLocation)
					fmt.Println("OK")
				}
			}
		}
	}

	//check that clamscan logs
	fmt.Print("    8. checking clamscan log for infected files: ")
	avLogLocation := filepath.Join(GetLog(SIP_SCAN_AV))
	f, err := os.Open(avLogLocation)
	if err != nil {
		fmt.Println("ERROR: could not open clamscan log")
		log.Println("[ERROR] could not open clamscan log")
	} else {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		infectedFiles := 0
		for scanner.Scan() {
			if !okPattern.MatchString(scanner.Text()) {
				infectedFiles++
				log.Println("[ERROR] found infected file:", scanner.Text())
			}
		}
		if infectedFiles > 0 {
			hasError = true
			fmt.Printf("ERROR, contains %d infected files\n", infectedFiles)
		} else {
			fmt.Println("OK")
		}
	}

	fmt.Print("    9. checking detox log: ")
	detoxStat, err := os.Stat(filepath.Join(GetLog(SIP_SCAN_DETOX)))
	if err != nil {
		hasWarning = true
		log.Println("[WARNING] no detox scan log available")
		fmt.Print("WARNING detox scan log missing\n")
	} else {
		if detoxStat.Size() > 0 {
			hasWarning = true
			log.Println("[WARNING] detox scan contains unresolved filename issues") // this can be fleshed out to print each individual detox errors
			fmt.Print("WARNING detox scan contains unresolved filename issues\n")
		} else {
			fmt.Print("OK\n")
		}
	}

	//finish up
	fmt.Printf("  * Validation report written to %s\n", logFile.Name())
	if hasError {
		fmt.Println("  * SIP HAS ERRORS")
	}

	if hasWarning {
		fmt.Println("  * SIP HAS WARNINGS")
	}

	if !hasError && !hasWarning {
		fmt.Println(" * NO ERRORS FOUND SIP IS VALID")
	}
	return nil
}

func ScanAV() ([]string, error) {
	fmt.Println("ewt sip scan av, ", VERSION)
	if err := loadConfig(); err != nil {
		return nil, err
	}

	logFilePath := GetLog(SIP_SCAN_AV)
	//create a logger and writer
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, err
	}
	defer logFile.Close()
	writer := bufio.NewWriter(logFile)
	defer writer.Flush()

	// collect all file paths before scanning
	var files []string
	if err := filepath.Walk(config.SIPLoc, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return nil, err
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		scanErr error
	)
	var errors = []string{}
	for _, path := range files {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			fmt.Println(" *  scanning", filepath.Join("sip", strings.ReplaceAll(p, config.SIPLoc, ""))) // get the directory name
			avCommand := exec.Command("clamdscan", "--fdpass", "--no-summary", p)                       // set the quarantine location in the config
			avOut, err := avCommand.CombinedOutput()
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				scanErr = fmt.Errorf("[ERROR] clamdscan malware detected: %s\n", p)
				errors = append(errors, scanErr.Error())
			}
			writer.Write(avOut)
		}(path)
	}

	wg.Wait()
	return errors, nil
}

func ScanDetox() error {
	fmt.Println("ewt sip scan detox,", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}

	logFilePath := GetLog(SIP_SCAN_DETOX)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return err
	}
	defer logFile.Close()
	writer := bufio.NewWriter(logFile)
	defer writer.Flush()
	var results = []string{}
	err = filepath.WalkDir(config.SIPLoc, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Run detox against this file only.
		result, err := detoxFile(path)
		if err != nil {
			return err
		}
		if result != nil {
			results = append(results, *result)
		}
		return nil
	})

	if len(results) > 0 {
		fmt.Println("  * detox chars found:")
		for _, result := range results {
			writer.WriteString(result)
			fmt.Println("   ", result)
		}
	} else {
		fmt.Println("  * no detox chars found")
	}

	return nil
}

func detoxFile(path string) (*string, error) {
	cmd := exec.Command("detox", "-n", path)

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if len(bytes.TrimSpace(out)) > 0 {
		result := string(out)
		return &result, nil
	}

	return nil, nil
}

func ScanDoubleExtensions() error {
	fmt.Println("ewt sip scan double extensions,", VERSION)
	if err := loadConfig(); err != nil {
		return err
	}
	logFilePath := GetLog(SIP_SCAN_EXTENSIONS)
	logFile, err := os.Create(logFilePath)
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	if err := filepath.Walk(config.SIPLoc, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			hasDoubleExtension(info.Name(), path)
		}

		return nil
	}); err != nil {
		return err
	}
	return nil
}

func hasDoubleExtension(filename string, path string) error {

	ext := filepath.Ext(filename)
	if ext == "" {
		return nil
	}

	base := strings.TrimSuffix(filename, ext)

	if filepath.Ext(base) == ext {
		fmt.Println(" * removing double extension from", path)
		dir := filepath.Dir(path)
		newPath := filepath.Join(dir, base)
		log.Printf("removing double extension from %s to %s", path, newPath)
		if err := os.Rename(path, newPath); err != nil {
			return err
		}
	}
	return nil
}

func woContains(s string, sl []string) bool {
	for _, sls := range sl {
		if s == sls {
			return true
		}
	}
	return false
}
