package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nyudlts/bytemath"
)

type DirectoryStats struct {
	Name           string
	Size           int64
	NumFiles       int
	NumDirectories int
}

func printDirectoryStats(pkgPath string) error {
	directoryStats := []DirectoryStats{}
	dirs, err := os.ReadDir(pkgPath)
	if err != nil {
		return err
	}

	for _, dir := range dirs {
		ds := DirectoryStats{dir.Name(), 0, 0, 0}
		if err := filepath.Walk(filepath.Join(pkgPath, dir.Name()), func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				ds.NumDirectories++
			} else {
				ds.NumFiles++
				ds.Size += info.Size()
			}
			return nil
		}); err != nil {
			return err
		}
		directoryStats = append(directoryStats, ds)
	}

	for _, ds := range directoryStats {
		fmt.Printf("  %s: %d files in %d directories, %s\n", ds.Name, ds.NumFiles, ds.NumDirectories, bytemath.ConvertBytesToHumanReadable(ds.Size))
	}

	return nil

}

func getPackageSize(pkgPath string) error {
	numFiles := 0
	numDirectories := 0
	sizeFiles := int64(0)

	if err := filepath.Walk(pkgPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			numDirectories++
		} else {
			numFiles++
			sizeFiles += info.Size()
		}
		return nil
	}); err != nil {
		return err
	}

	fmt.Printf("%s: %d files in %d directories, %s\n", pkgPath, numFiles, numDirectories, bytemath.ConvertBytesToHumanReadable(sizeFiles))
	return nil
}
