package lib

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var (
	collectionCode string
	sourceLoc      string
	configLoc      string
	gzipFile       *os.File
	gzipWriter     *gzip.Writer
	tarWriter      *tar.Writer
)

func InitProject(cCode string, sLoc string, config string) error {
	fmt.Println("ewt project init, version", VERSION)

	//determine config location
	if config == "" {
		user, err := user.Current()
		if err != nil {
			panic(fmt.Sprintf("Error getting current user: %v", err))
		}
		if runtime.GOOS == "windows" {
			hd := os.Getenv("USERPROFILE")
			configLoc = filepath.Join(hd, ".config", "ewt.config")
		}
		configLoc = filepath.Join(user.HomeDir, ".config", "ewt.config")
	} else {
		configLoc = config
	}

	collectionCode = cCode
	sourceLoc = sLoc

	//generate ewt config
	if err := generateConfig(); err != nil {
		return err
	}

	//make project directory
	if err := mkProjectDir(); err != nil {
		return err
	}

	//write the ewt-config to the project directory
	if err := writeEWTConfig(); err != nil {
		return err
	}

	return nil
}

func generateConfig() error {
	fmt.Println("  * generating ewt config")

	//read the initial file
	configBytes, err := os.ReadFile(configLoc)
	if err != nil {
		return err
	}

	//unmarshal to config options
	config = Config{}
	if err := json.Unmarshal(configBytes, &config); err != nil {
		return err
	}

	//add config members
	config.PartnerCode = strings.Split(collectionCode, "_")[0]
	config.CollectionCode = collectionCode
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	config.ProjectLoc = filepath.Join(wd, collectionCode)
	config.SIPLoc = filepath.Join(config.ProjectLoc, "sip")
	config.AIPLoc = filepath.Join(config.ProjectLoc, "aips")
	config.LogLoc = filepath.Join(config.ProjectLoc, "logs")
	config.XferLoc = filepath.Join(config.ProjectLoc, "xfer")
	config.SourceLoc, err = filepath.Abs(sourceLoc)
	switch runtime.GOOS {
	case "linux":
		{
			if !strings.HasSuffix(config.SourceLoc, "/") {
				//rsync on linux needs the trailing slash to copy contents of directory
				config.SourceLoc += "/"
			}
		}
	case "windows":
		{
			//ensure windows paths use backslashes
			config.SourceLoc = filepath.Clean(config.SourceLoc)
			if !strings.HasSuffix(config.SourceLoc, "\\") {
				config.SourceLoc += "\\"
			}
		}
	}
	config.WorkLoc = filepath.Join(config.ProjectLoc, "aips", "aip_queue")

	return nil
}

func mkProjectDir() error {
	fmt.Println("  * generating ewt project directory")

	//create the project directory
	if err := os.Mkdir(config.ProjectLoc, 0775); err != nil {
		return err
	}

	//create the aips directories
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips", "in"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips", "valid"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips", "complete"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips", "failed"), 0775); err != nil {
		return err
	}

	//create the aip_queue working directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips", "aip_queue"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips", "aip_queue", "in"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips", "aip_queue", "failure"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips", "aip_queue", "complete"), 0775); err != nil {
		return err
	}

	//create the logs directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "logs"), 0775); err != nil {
		return err
	}

	//create the resync output directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "logs", "rsync"), 0775); err != nil {
		return err
	}

	//create the resync output directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "logs", "fix"), 0775); err != nil {
		return err
	}

	//create the sip output directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "sip"), 0775); err != nil {
		return err
	}

	//create the xfer directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "xfer"), 0775); err != nil {
		return err
	}

	return nil
}

func writeEWTConfig() error {

	fmt.Println("  * writing ewt config to project directory")

	//marshall the updated config
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	//write the config to the project directory
	if err := os.WriteFile(filepath.Join(config.ProjectLoc, "config.json"), b, 0755); err != nil {
		return err
	}

	return nil
}

func ArchiveProject(projectLoc string) error {
	fmt.Println("ewt project archive, version", VERSION)

	//check that the project location contains a config file
	if err := loadConfigPath(filepath.Join(projectLoc, "config.json")); err != nil {
		return fmt.Errorf("error loading config from project location: %v", err)
	}

	logPath := GetLog(PROJECT_ARCHIVE_CREATE)
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("error creating log file: %v", err)
	}

	log.SetOutput(logFile)
	log.Printf("[INFO] starting project archive for %s\n", config.CollectionCode)

	// Remove AIP Directory
	fmt.Println("  * removing aips directory")
	aipsDir := filepath.Join(projectLoc, "aips")
	log.Println("[INFO] removing aips directory")
	if err := removeDirectory(aipsDir); err != nil {
		log.Printf("[ERROR] error removing aips directory: %v", err)
		fmt.Printf("  * error removing aips directory: %v\n", err)
		return (err)
	}

	// Remove XferDIrectory
	fmt.Println("  * removing xfer directory")
	log.Println("[INFO] removing xfer directory")
	xferDir := filepath.Join(projectLoc, "xfer")
	if err := removeDirectory(xferDir); err != nil {
		log.Printf("[ERROR] error removing xfer directory: %v", err)
		fmt.Printf("  * error removing xfer directory: %v\n", err)
		return (err)
	}

	//move metadata directory to project root
	fmt.Println("  * moving metadata directory to project root")
	log.Println("[INFO] moving metadata directory to project root")
	metadataDir := filepath.Join(projectLoc, "sip", "metadata")
	if err := os.Rename(metadataDir, filepath.Join(projectLoc, "metadata")); err != nil {
		log.Printf("[ERROR] error moving metadata directory to project root: %v", err)
		fmt.Printf("  * error moving metadata directory to project root: %v\n", err)
		return (err)
	}

	//remmove sip directory
	fmt.Println("  * removing sip directory")
	log.Println("[INFO] removing sip directory")
	sipDir := filepath.Join(projectLoc, "sip")
	if err := removeDirectory(sipDir); err != nil {
		log.Printf("[ERROR] error removing sip directory: %v", err)
		fmt.Printf("  * error removing sip directory: %v\n", err)
		return (err)
	}

	log.Println("[INFO] creating gzip of project directory")
	fmt.Println("  * creating gzip of project directory")
	if _, err := os.Stat(config.ProjectLoc); err != nil {
		return err
	}

	//create the gzip file
	timestamp := time.Now().Format("20060102-150405")
	projectName := filepath.Base(config.ProjectLoc)
	gzipName := filepath.Join("completed", fmt.Sprintf("%s-%s.tgz", projectName, timestamp))
	log.Printf("[INFO] creating gzip file: %s\n", gzipName)
	fmt.Printf("  * creating gzip file: %s\n", gzipName)
	gzipFile, err = os.Create(gzipName)
	if err != nil {
		log.Printf("[ERROR] error creating gzip file: %v\n", err)
		fmt.Printf("  * error creating gzip file: %v\n", err)
		return err
	}
	defer gzipFile.Close()

	//create the gzip writer
	gzipWriter = gzip.NewWriter(gzipFile)
	defer gzipWriter.Close()

	//create the tar writer
	tarWriter = tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Create a gzip of the project
	fmt.Println("  * compressing project directory")
	log.Println("[INFO] compressing project directory")
	if err := createGzip(); err != nil {
		fmt.Printf("  * error compressing project directory: %v\n", err)
		log.Printf("[ERROR] error compressing project directory: %v", err)
		return (err)
	}

	logFile.Close()

	fmt.Println("  * adding archive create log file to tar")
	logFileInfo, err := os.Stat(logPath)
	if err != nil {
		return fmt.Errorf("error stating archive createlog file: %v", err)
	}
	if err := writeToTar(logPath, logFileInfo, filepath.Base(config.ProjectLoc)); err != nil {
		return fmt.Errorf("error adding archive create log file to tar: %v", err)
	}

	// Remove the project directory
	fmt.Println("  * removing project directory")
	if err := os.RemoveAll(projectLoc); err != nil {
		fmt.Printf("  * error removing project directory: %v\n", err)
		return (err)
	}

	return nil
}

var archiveLogPtn = regexp.MustCompile(".*project-archive-create.log$")

// Derived from: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func createGzip() error {
	fmt.Println("  * creating gzip of project directory")

	//walk the project location and add all files to the tar
	return filepath.Walk(config.ProjectLoc, func(file string, fi os.FileInfo, err error) error {
		//return an error
		if err != nil {
			return err
		}

		//return if the file is not regular
		if !fi.Mode().IsRegular() {
			return nil
		}

		if archiveLogPtn.MatchString(fi.Name()) {
			log.Printf("[INFO] skipping log file: %s\n", fi.Name())
			fmt.Printf("  * skipping log file: %s\n", fi.Name())
			return nil
		}

		if err := writeToTar(file, fi, filepath.Base(config.ProjectLoc)); err != nil {
			return err
		}

		return nil
	})
}

func writeToTar(file string, fi os.FileInfo, projectName string) error {
	log.Printf("[INFO] adding file to tar: %s\n", file)
	fmt.Printf("  * adding file to tar: %s\n", file)

	//create a new tar header
	header, err := tar.FileInfoHeader(fi, fi.Name())
	if err != nil {
		return err
	}

	relPath, err := filepath.Rel(config.ProjectLoc, file)
	if err != nil {
		return err
	}

	header.Name = filepath.ToSlash(filepath.Join(projectName, relPath))

	//write the header
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}

	//read the file
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	//copy the file data to the tar

	if _, err := io.Copy(tarWriter, f); err != nil {
		log.Printf("[ERROR] error adding file to tar: %v\n", err)
		fmt.Printf("  * error adding %v to tar: %v\n", file, err)
		return err
	}

	//close the file
	f.Close()

	return nil

}

func removeDirectory(dir string) error {
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			fmt.Println("    * removing file: ", path)
			log.Printf("[INFO] removing file: %s\n", path)
			if err := os.Remove(path); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := os.RemoveAll(dir); err != nil {
		return err
	}

	return nil
}

func VerifyProjectArchive(archivePath string) error {
	fmt.Println("ewt project archive verify, version", VERSION)
	return nil
}
