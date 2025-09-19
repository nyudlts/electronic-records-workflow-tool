package lib

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	amatica "github.com/nyudlts/go-archivematica"
)

const timeFormat = "2006-01-02 15:04:05"

var (
	locationName     string
	polltime         int
	poll             time.Duration
	client           *amatica.AMClient
	amaticaConfigLoc string
	xferDirs         []os.DirEntry
	aipWriter        *bufio.Writer
	amLocation       amatica.Location
)

func TransferToArchivematica(p int, configLoc string) error {
	polltime = p
	amaticaConfigLoc = configLoc
	fmt.Println("ewt amatica transfer, version", VERSION)
	//load configuration file
	if err := loadConfig(); err != nil {
		return err
	}

	//move this to a func
	//create a log file
	logFilename := filepath.Join(config.LogLoc, fmt.Sprintf("%s-amatica-transfer.log", config.CollectionCode))

	logFile, err := os.Create(logFilename)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//create the aip-file and writer
	fmt.Printf("  * creating %s-aip-file.txt\n", config.CollectionCode)
	log.Printf("[INFO] creating %s-aip-file.txt", config.CollectionCode)
	of, err := os.Create(filepath.Join(config.LogLoc, fmt.Sprintf("%s-aip-file.txt", config.CollectionCode)))
	if err != nil {
		panic(err)
	}
	defer of.Close()
	aipWriter = bufio.NewWriter(of)

	//check flags
	if err := checkFlags(); err != nil {
		return err
	}

	//setup client
	if err := setupClient(); err != nil {
		return err
	}

	if err := transferDirectories(); err != nil {
		return err
	}

	return nil
}

func checkFlags() error {

	//check config exists
	if amaticaConfigLoc != "" {
		fi, err := os.Stat(amaticaConfigLoc)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return fmt.Errorf("%s is a directory, config file required", amaticaConfigLoc)
		}
	} else {
		currentUser, err := user.Current()
		if err != nil {
			return (err)
		}

		if runtime.GOOS == "windows" {
			cu := strings.Split(currentUser.Username, "\\")[1]
			amaticaConfigLoc = fmt.Sprintf("C:\\Users\\%s\\.config\\go-archivematica.yml", cu)
		} else {
			amaticaConfigLoc = fmt.Sprintf("/home/%s/.config/go-archivematica.yml", currentUser.Username)
		}

		client, err = amatica.NewAMClient(amaticaConfigLoc, 20)
		if err != nil {
			return err
		}
	}

	return nil
}

func setupClient() error {
	// set the transfer location
	locationName = config.AMTransferSource

	//set the poll time
	fmt.Printf("  * setting polling time to %d seconds\n", polltime)
	log.Printf("[INFO] setting polling time to %d seconds", polltime)
	poll = time.Duration(polltime * int(time.Second))
	//create a client
	fmt.Println("  * creating go-archivematica client")
	log.Println("[INFO] creating go-archivematica client")
	var err error
	client, err = amatica.NewAMClient(amaticaConfigLoc, 20)
	if err != nil {
		return err
	}

	//process the directory
	fmt.Printf("  * reading source directory: %s\n", "xfer/")
	log.Printf("[INFO] reading source directory: %s", "xfer/")
	xferDirs, err = os.ReadDir("xfer")
	if err != nil {
		return err
	}

	if len(xferDirs) < 1 {
		return fmt.Errorf("transfer directory is empty")
	}

	return nil
}

func transferDirectories() error {
	fmt.Printf("transferring packages from %s\n", "xfer/")
	log.Printf("[INFO] transferring packages from %s", "xfer")

	for _, xferDir := range xferDirs {
		xipPath := filepath.Join(config.CollectionCode, "xfer", xferDir.Name())
		if err := transferPackage(xipPath); err != nil {
			//log the err instead
			return err
		}
	}

	return nil
}

func transferPackage(xipPath string) error {

	//fix this to work with windowa paths...
	//initialize the transfer
	xipName := filepath.Base(xipPath)
	fmt.Printf("\ninitializing transfer for %s\n", xipName)
	amXIPPath, err := initTransfer(xipPath)
	if err != nil {
		return err
	}
	fmt.Printf("transfer %s initialized\n", amXIPPath)
	log.Printf("[INFO] transfer %s initialized\n", amXIPPath)

	//request the transfer through archivematica
	fmt.Printf("requesting transfer processing for %s\n", xipName)
	transferUUID, err := requestTransfer(amXIPPath)
	if err != nil {
		return err
	}
	fmt.Printf("transfer processing requested for %s-%s\n", amXIPPath, transferUUID)
	log.Printf("[INFO] transfer processing requested for %s-%s", amXIPPath, transferUUID)

	//approve the transfer
	fmt.Printf("approving %s: %s for transfer processing\n", amXIPPath, transferUUID)
	transferStatus, err := approveTransfer(transferUUID)
	if err != nil {
		return err
	}

	xferLabel := fmt.Sprintf("%s-%s", filepath.Base(amXIPPath), transferUUID)
	fmt.Printf("transfer processing approved for %s\n", xferLabel)
	log.Printf("[INFO] transfer processing archivematica approved for %s", xferLabel)

	//transfer processing
	fmt.Printf("transfer processing started for %s\n", xferLabel)
	transferStatus, err = transferProcessing(transferStatus.UUID.String())
	if err != nil {
		return err
	}
	fmt.Printf("transfer processing completed for %s\n", xferLabel)
	log.Printf("[INFO] transfer processing completed for %s", xferLabel)

	//ingest processing
	ingestLabel := fmt.Sprintf("%s-%s", filepath.Base(amXIPPath), transferStatus.SIPUUID)
	fmt.Printf("ingest processing started for %s\n", ingestLabel)
	//pause for api to update
	time.Sleep(5 * time.Second)
	ingestStatus, err := ingestProcessing(transferStatus.SIPUUID)
	if err != nil {
		return err
	}
	fmt.Printf("ingest processing completed for %s\n", ingestLabel)
	log.Printf("[INFO] ingest processing completed for %s", ingestLabel)

	//write path to aip-file
	aipPath, err := amatica.ConvertUUIDToAMDirectory(ingestStatus.UUID.String())
	if err != nil {
		return err
	}

	aipPath = filepath.Join(aipPath, fmt.Sprintf("%s-%s", filepath.Base(xipPath), ingestStatus.UUID.String()))

	aipPath = fmt.Sprintf("%s%s", "/mnt/amatica/AIPsStore/", aipPath)
	fmt.Printf("writing %s to aip-file\n", aipPath)
	aipWriter.WriteString(fmt.Sprintf("%s\n", aipPath))
	aipWriter.Flush()
	log.Printf("[INFO] %s written to aip-file", aipPath)
	fmt.Printf("%s written to aip-file\n", aipPath)

	//done
	return nil
}

func initTransfer(xipPath string) (string, error) {
	var err error
	amLocation, err = client.GetLocationByName(locationName)
	if err != nil {
		return "", err
	}

	amXIPPath := filepath.Join(amLocation.Path, xipPath)

	return amXIPPath, nil
}

func requestTransfer(xipPath string) (string, error) {
	startTransferResponse, err := client.StartTransfer(amLocation.UUID, xipPath)
	if err != nil {
		return "", err
	}

	//catch the soft error
	if regexp.MustCompile("^Error").MatchString(startTransferResponse.Message) {
		return "", fmt.Errorf("%s", startTransferResponse.Message)
	}

	fmt.Printf("transfer request message: %s\n", startTransferResponse.Message)
	log.Printf("[INFO] transfer request message: %s", startTransferResponse.Message)

	//get the uuid for the transfer
	uuid, err := startTransferResponse.GetUUID()
	if err != nil {
		return "", err
	}
	return uuid, nil
}

func approveTransfer(xferUUID string) (amatica.TransferStatus, error) {
	foundUnapproved := false
	for !foundUnapproved {
		var err error
		foundUnapproved, err = findUnapprovedTransfer(xferUUID)
		if err != nil {
			return amatica.TransferStatus{}, err
		}

		if !foundUnapproved {
			fmt.Printf("  * %s waiting for approval process to complete\n", time.Now().Format(timeFormat))
			time.Sleep(poll)
		}
	}

	//approve the transfer
	transfer, err := client.GetTransferStatus(xferUUID)
	if err != nil {
		return amatica.TransferStatus{}, err
	}

	if err := client.ApproveTransfer(transfer.Directory, "standard"); err != nil {
		return amatica.TransferStatus{}, err
	}

	approvedTransfer, err := client.GetTransferStatus(xferUUID)
	if err != nil {
		return amatica.TransferStatus{}, err
	}

	return approvedTransfer, nil
}

func transferProcessing(xferUUID string) (amatica.TransferStatus, error) {

	//change this logic over to a channel
	foundCompleted := false
	for !foundCompleted {
		ts, err := client.GetTransferStatus(xferUUID)
		if err != nil {
			return amatica.TransferStatus{}, err
		}

		if ts.Status == "FAILED" {
			return amatica.TransferStatus{}, fmt.Errorf("%s", ts.Microservice)
		}

		if ts.Status == "" {
			return amatica.TransferStatus{}, fmt.Errorf("no status being returned")
		}

		if ts.Status == "COMPLETE" {
			foundCompleted = true
		}

		if !foundCompleted {
			fmt.Printf("  * %s Transfer Status: %s,  Microservice: %s\n", time.Now().Format(timeFormat), ts.Status, ts.Microservice)
			time.Sleep(poll)
		}
	}

	completedTransfer, err := client.GetTransferStatus(xferUUID)
	if err != nil {
		return amatica.TransferStatus{}, err
	}

	sipUUID := completedTransfer.SIPUUID
	if sipUUID == "" {
		return amatica.TransferStatus{}, fmt.Errorf("no sip-uuid returned")
	}

	return completedTransfer, nil
}

func ingestProcessing(ingestUUID string) (amatica.IngestStatus, error) {
	foundIngestCompleted := false
	var ingestStatus amatica.IngestStatus
	var err error
	for !foundIngestCompleted {
		ingestStatus, err = client.GetIngestStatus(ingestUUID)
		if err != nil {
			return amatica.IngestStatus{}, err
		}

		if ingestStatus.Status == "FAILED" {
			return amatica.IngestStatus{}, fmt.Errorf("%s", ingestStatus.Microservice)
		}

		if ingestStatus.Status == "" {
			return amatica.IngestStatus{}, fmt.Errorf("no status being returned")
		}

		if ingestStatus.Status == "COMPLETE" {
			foundIngestCompleted = true
		}

		if !foundIngestCompleted {
			fmt.Printf("  * %s Ingest Status: %s,  Microservice: %s\n", time.Now().Format(timeFormat), ingestStatus.Status, ingestStatus.Microservice)
			time.Sleep(poll)
		}
	}

	return ingestStatus, nil

}

func findUnapprovedTransfer(uuid string) (bool, error) {
	unapprovedTransfers, err := client.GetUnapprovedTransfers()
	if err != nil {
		return false, err
	}

	unapprovedTransfersMap, err := client.GetUnapprovedTransfersMap(unapprovedTransfers)
	if err != nil {
		return false, err
	}

	//find the unapproved transfer
	for k := range unapprovedTransfersMap {
		if k == uuid {
			return true, nil
		}
	}

	return false, nil
}
