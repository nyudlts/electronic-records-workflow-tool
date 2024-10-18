package cmd

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	amatica "github.com/nyudlts/go-archivematica"
	"github.com/spf13/cobra"
)

var (
	poll             time.Duration
	client           *amatica.AMClient
	xferDirs         []fs.DirEntry
	xferDirectoryPtn *regexp.Regexp
)

const locationName = "amatica rws ingest point"

func init() {
	if runtime.GOOS == "windows" {
		windows = true
	}
	xferAmaticaCmd.Flags().StringVar(&amaticaConfigLoc, "config", "", "")
	xferAmaticaCmd.Flags().StringVar(&xferDirectory, "transfer-directory", "", "")
	xferAmaticaCmd.Flags().IntVar(&pollTime, "poll", 5, "")
	xferAmaticaCmd.Flags().StringVar(&ersRegex, "regexp", "", "")
	rootCmd.AddCommand(xferAmaticaCmd)
}

var xferAmaticaCmd = &cobra.Command{
	Use: "transfer-amatica",
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkFlags(); err != nil {
			panic(err)
		}

		if err := setup(); err != nil {
			panic(err)
		}

		if err := xferDiretories(); err != nil {
			panic(err)
		}
	},
}

func checkFlags() error {
	//check config exists
	fi, err := os.Stat(amaticaConfigLoc)
	if err != nil {
		return err
	}
	if fi.IsDir() {
		return fmt.Errorf("%s is a directory, config file required", amaticaConfigLoc)
	}

	//check transfer directory exists
	fi, err = os.Stat(xferDirectory)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("%s is not a directory", xferDirectory)
	}
	//check regexp is not empty
	if ersRegex == "" {
		return fmt.Errorf("regexp is empty, must be defined")
	}
	xferDirectoryPtn = regexp.MustCompile(ersRegex)

	return nil
}

func setup() error {
	fmt.Println("Creating Log File")
	logFile, err := os.Create("am-tools-transfer.log")
	if err != nil {
		return err
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	//create an output file
	fmt.Println("creating aip-file.txt")
	log.Println("[INFO] creating aip-file.txt")
	of, err := os.Create("aip-file.txt")
	if err != nil {
		return err
	}
	defer of.Close()
	writer = bufio.NewWriter(of)

	//set the poll time
	fmt.Printf("setting polling time to %d seconds\n", pollTime)
	log.Printf("[INFO] setting polling time to %d seconds", pollTime)
	poll = time.Duration(pollTime)

	//create a client
	fmt.Println("creating go-archivematica client")
	log.Println("[INFO] creating go-archivematica client")
	client, err = amatica.NewAMClient(amaticaConfigLoc, 20)
	if err != nil {
		return err
	}

	//process the directory
	fmt.Printf("Reading source directory: %s\n", xferDirectory)
	log.Printf("[INFO] reading source directory: %s", xferDirectory)

	xferDirs, err = os.ReadDir(xferDirectory)
	if err != nil {
		panic(err)
	}

	return nil
}

func xferDiretories() error {
	fmt.Printf("Transferring packages from %s\n", xferDirectory)
	log.Printf("[INFO] transferring packages from %s", xferDirectory)

	for _, xferDir := range xferDirs {
		if xferDirectoryPtn.MatchString(xferDir.Name()) {
			xipPath := filepath.Join(xferDirectory, xferDir.Name())
			if err := transferPackage(xipPath); err != nil {
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
	xipName := filepath.Base(xipPath)

	location, err := client.GetLocationByName(locationName)
	if err != nil {
		return err
	}

	fmt.Println("Location:", location)

	amXIPPath := filepath.Join(location.Path, xipName)
	if windows {
		amXIPPath = strings.Replace(amXIPPath, "\\", "/", -1)
	}
	fmt.Printf("\nTransfering package: %s to archivematica\n", amXIPPath)
	log.Printf("[INFO] transfering package: %s to archivematica", amXIPPath)
	return nil
}
