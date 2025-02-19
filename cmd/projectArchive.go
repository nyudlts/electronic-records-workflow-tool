package cmd

import (
	"archive/tar"
	"compress/gzip"
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
	Use: "archive",
	Run: func(cmd *cobra.Command, args []string) {
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

		// Move the gzip to the archive directory
		if err := os.Rename(projectLocation+".tgz", filepath.Join("completed", projectLocation+".tgz")); err != nil {
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

	gzipFile, err := os.Create(projectLocation + ".tgz")
	if err != nil {
		return err
	}
	defer gzipFile.Close()

	gzipWriter := gzip.NewWriter(gzipFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(projectLocation, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(strings.Replace(file, projectLocation, "", -1), string(filepath.Separator))

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		if _, err := io.Copy(tarWriter, f); err != nil {
			return err
		}

		f.Close()

		return nil
	})
}
