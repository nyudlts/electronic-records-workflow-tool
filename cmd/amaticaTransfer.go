package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	amatica "github.com/nyudlts/go-archivematica"
	"github.com/spf13/cobra"
)

const (
	locationName = "Default transfer source"
	timeFormat   = "2006-01-02 15:04:05"
)

var (
	poll             time.Duration
	client           *amatica.AMClient
	xferDirs         []fs.DirEntry
	xferDirectoryPtn *regexp.Regexp
	aipWriter        *bufio.Writer
	amLocation       amatica.Location
)

func init() {
	fmt.Printf("adoc %s transfer-am", version)
	if runtime.GOOS == "windows" {
		fmt.Println("setting Windows mode")
		windows = true
	}
	xferAmaticaCmd.Flags().StringVar(&amaticaConfigLoc, "config", "", "")
	xferAmaticaCmd.Flags().StringVar(&xferDirectory, "xfer-directory", "", "")
	xferAmaticaCmd.Flags().IntVar(&pollTime, "poll", 5, "")
	xferAmaticaCmd.Flags().StringVar(&ersRegex, "regexp", "", "")
	rootCmd.AddCommand(xferAmaticaCmd)
}

var xferAmaticaCmd = &cobra.Command{
	Use:   "transfer-am",
	Short: "Transfer SIPS to R*",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("checking program flags")
		if err := checkFlags(); err != nil {
			panic(err)
		}

		fmt.Println("creating log File")
		logFile, err := os.Create("am-tools-transfer.log")
		if err != nil {
			panic(err)
		}
		defer logFile.Close()
		log.SetOutput(logFile)

		//create an output file
		fmt.Println("creating aip-file.txt")
		log.Println("[INFO] creating aip-file.txt")
		of, err := os.Create("aip-file.txt")
		if err != nil {
			panic(err)
		}
		defer of.Close()
		aipWriter = bufio.NewWriter(of)

		if err := setup(); err != nil {
			panic(err)
		}

		if err := xferDirectories(); err != nil {
			panic(err)
		}
	},
}

func checkFlags() error {
	//check config exists					//modification check $HOME/.config/go-archivematica if not defined
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

		configPath := fmt.Sprintf("/home/%s/.config/go-archivematica.yml", currentUser.Username)
		cf, err := os.Stat(configPath)
		if err != nil {
			return err
		}

		if cf.IsDir() {
			return fmt.Errorf("%s is a directory, config file required", configPath)
		}
		
		amaticaConfigLoc = configPath
	}

	//check transfer directory exists
	fi, err = os.Stat(xferDirectory)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", xferDirectory)
	}

		
	}

	//check regexp is not empty
	if ersRegex == "" {
		return fmt.Errorf("regexp is empty, must be defined")
	}
	xferDirectoryPtn = regexp.MustCompile(ersRegex)

	return nil
}

func setup() error {

	//set the poll time
	fmt.Printf("setting polling time to %d seconds\n", pollTime)
	log.Printf("[INFO] setting polling time to %d seconds", pollTime)
	poll = time.Duration(pollTime * int(time.Second))
	//create a client
	fmt.Println("creating go-archivematica client")
	log.Println("[INFO] creating go-archivematica client")
	var err error
	client, err = amatica.NewAMClient(amaticaConfigLoc, 20)
	if err != nil {
		return err
	}

	//process the directory
	fmt.Printf("reading source directory: %s\n", xferDirectory)
	log.Printf("[INFO] reading source directory: %s", xferDirectory)
	xferDirs, err = os.ReadDir(xferDirectory)
	if err != nil {
		return err
	}

	if len(xferDirs) < 1 {
		return fmt.Errorf("transfer directory is empty")
	}

	return nil
}

func xferDirectories() error {
	fmt.Printf("transferring packages from %s\n", xferDirectory)
	log.Printf("[INFO] transferring packages from %s", xferDirectory)

	for _, xferDir := range xferDirs {
		if xferDirectoryPtn.MatchString(xferDir.Name()) {
			xipPath := filepath.Join(xferDirectory, xferDir.Name())
			if err := transferPackage(xipPath); err != nil {
				//log the err instead
				return err
			}
		} else {
			fmt.Printf("skipping %s, does not match pattern %s", xferDir.Name(), ersRegex)
			continue
		}
	}

	return nil
}

func transferPackage(xipPath string) error {

	//initialize the transfer
	xipName := filepath.Base(xipPath)
	fmt.Printf("initializing transfer for %s\n", xipName)
	amXIPPath, err := initTransfer(xipName)
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
	if windows {
		aipPath = strings.Replace(aipPath, "\\", "/", -1)
	}

	aipPath = fmt.Sprintf("%s%s", "/mnt/amatica/AIPsStore/", aipPath)
	fmt.Printf("writing %s to aip-file\n", aipPath)
	aipWriter.WriteString(fmt.Sprintf("%s\n", aipPath))
	aipWriter.Flush()
	log.Printf("[INFO] %s written to aip-file", aipPath)
	fmt.Printf("%s written to aip-file\n", aipPath)

	//done
	return nil
}

func initTransfer(xipName string) (string, error) {
	var err error
	amLocation, err = client.GetLocationByName(locationName)
	if err != nil {
		return "", err
	}

	//convert windows path separators
	amXIPPath := filepath.Join(amLocation.Path, xipName)
	if windows {
		amXIPPath = strings.Replace(amXIPPath, "\\", "/", -1)
	}
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
	for k, _ := range unapprovedTransfersMap {
		if k == uuid {
			return true, nil
		}
	}

	return false, nil
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
			return amatica.TransferStatus{}, fmt.Errorf(ts.Microservice)
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
			return amatica.IngestStatus{}, fmt.Errorf(ingestStatus.Microservice)
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
