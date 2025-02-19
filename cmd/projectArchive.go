package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var projectLocation string

func init() {
	archiveCmd.Flags().StringVarP(&projectLocation, "project-location", "p", "", "Project name")
	projectCmd.AddCommand(archiveCmd)
}

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive a project",
	Run: func(cmd *cobra.Command, args []string) {
		// Add a warning message

		// Remove AIP Directory
		aipsDir := filepath.Join(projectLocation, "aips")
		if err := os.RemoveAll(aipsDir); err != nil {
			panic(err)
		}

		// Remove XferDIrectory
		xferDir := filepath.Join(projectLocation, "xfer")
		if err := os.RemoveAll(xferDir); err != nil {
			panic(err)
		}

		// Create a gzip of the project
		if err := createGzip(); err != nil {
			panic(err)
		}

		// Remove the project directory
		if err := os.RemoveAll(projectLocation); err != nil {
			panic(err)
		}
	},
}

// Derived from: https://medium.com/@skdomino/taring-untaring-files-in-go-6b07cf56bc07
func createGzip() error {
	if _, err := os.Stat(projectLocation); err != nil {
		return err
	}

	//create the gzip file
	gzipName := filepath.Join("completed", fmt.Sprintf("%s.tgz", projectLocation))
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
	return filepath.Walk(projectLocation, func(file string, fi os.FileInfo, err error) error {
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
		header.Name = strings.TrimPrefix(strings.Replace(file, projectLocation, "", -1), string(filepath.Separator))

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
