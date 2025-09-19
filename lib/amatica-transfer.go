package lib

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	amatica "github.com/nyudlts/go-archivematica"
)

var (
	locationName     string
	polltime         int
	poll             time.Duration
	client           *amatica.AMClient
	amaticaConfigLoc string
	xferDirs         []os.DirEntry
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

	//create the aip-file
	fmt.Printf("  * creating %s-aip-file.txt\n", config.CollectionCode)
	log.Printf("[INFO] creating %s-aip-file.txt", config.CollectionCode)
	of, err := os.Create(filepath.Join(config.LogLoc, fmt.Sprintf("%s-aip-file.txt", config.CollectionCode)))
	if err != nil {
		panic(err)
	}
	defer of.Close()

	//check flags
	if err := checkFlags(); err != nil {
		return err
	}

	//setup client
	if err := setupClient(); err != nil {
		return err
	}

	fmt.Println(client)

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
