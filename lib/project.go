package lib

import (
	"archive/tar"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

//go:embed ewt-config.yml
var vfs embed.FS

var (
	collectionCode string
	sourceLoc      string
	projectLoc     string
)

func InitProject(cCode string, sLoc string) error {
	fmt.Println("ewt project init, versions", VERSION)

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
	configBytes, err := vfs.ReadFile("ewt-config.yml")
	if err != nil {
		return err
	}

	//unmarshal to config options
	config = Config{}
	if err := yaml.Unmarshal(configBytes, &config); err != nil {
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
	if err != nil {
		return err
	}

	return nil
}

func mkProjectDir() error {
	fmt.Println("  * generating ewt project directory")

	//create the project directory
	if err := os.Mkdir(config.ProjectLoc, 0775); err != nil {
		return err
	}

	//create the aips directory
	if err := os.Mkdir(filepath.Join(config.ProjectLoc, "aips"), 0775); err != nil {
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
	b, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	//write the config to the project directory
	if err := os.WriteFile(filepath.Join(config.ProjectLoc, "config.yml"), b, 0755); err != nil {
		return err
	}

	return nil
}

func ArchiveProject(pl string) error {
	fmt.Println("ewt project archive, version", VERSION)
	projectLoc = pl

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
	gzipName := filepath.Join("completed", fmt.Sprintf("%s.tgz", projectLoc))
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
