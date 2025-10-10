package lib

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	collectionCode string
	sourceLoc      string
	projectLoc     string
	configLoc      string
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
	config.WorkLoc = filepath.Join(config.LogLoc, "aip_queue")

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

	//create the logs directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "logs"), 0775); err != nil {
		return err
	}

	//create the resync output directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "logs", "rsync"), 0775); err != nil {
		return err
	}

	//create the aip_queue working directory
	if err := os.Mkdir(filepath.Join(config.LogLoc, "aip_queue"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.LogLoc, "aip_queue", "in"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.LogLoc, "aip_queue", "failure"), 0775); err != nil {
		return err
	}

	if err := os.Mkdir(filepath.Join(config.LogLoc, "aip_queue", "success"), 0775); err != nil {
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

func ArchiveProject(pl string) error {
	fmt.Println("ewt project archive, version", VERSION)
	projectLoc = pl

	//check that the project location contains a config file
	if err := loadConfigPath(filepath.Join(pl, "config.json")); err != nil {
		return fmt.Errorf("error loading config from project location: %v", err)
	}

	// Remove AIP Directory
	fmt.Println("  * removing aips directory")
	aipsDir := filepath.Join(projectLoc, "aips")
	if err := os.RemoveAll(aipsDir); err != nil {
		panic(err)
	}

	// Remove XferDIrectory
	fmt.Println("  * removing xfer directory")
	xferDir := filepath.Join(projectLoc, "xfer")
	if err := os.RemoveAll(xferDir); err != nil {
		panic(err)
	}

	// Create a gzip of the project
	fmt.Println("  * compressing project directory")
	if err := createGzip(); err != nil {
		panic(err)
	}

	// Remove the project directory
	fmt.Println("  * removing project directory")
	if err := os.RemoveAll(projectLoc); err != nil {
		panic(err)
	}

	return nil
}

// Derived from: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func createGzip() error {
	if _, err := os.Stat(projectLoc); err != nil {
		return err
	}

	//create the gzip file
	timestamp := time.Now().Format("20060102-150405")
	gzipName := filepath.Join("completed", fmt.Sprintf("%s-%s.tgz", projectLoc, timestamp))
	gzipFile, err := os.Create(gzipName)
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	//create the gzip writer
	gzipWriter := gzip.NewWriter(gzipFile)
	defer gzipWriter.Close()

	//create the tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	//walk the project location and add all files to the tar
	return filepath.Walk(projectLoc, func(file string, fi os.FileInfo, err error) error {
		//return an error
		if err != nil {
			return err
		}

		//return if the file is not regular
		if !fi.Mode().IsRegular() {
			return nil
		}

		//create a new tar header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}
		header.Name = strings.TrimPrefix(strings.Replace(file, projectLoc, "", -1), string(filepath.Separator))

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
			return err
		}

		//close the file
		f.Close()

		return nil
	})
}
