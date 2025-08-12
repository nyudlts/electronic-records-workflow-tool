package lib

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func CleanSip() error {
	fmt.Println("executing sip clean")
	//load the project configuration
	if err := loadConfig(); err != nil {
		return err
	}
	deleteCount := 0
	if err := filepath.Walk(config.SIPLoc, func(path string, info fs.FileInfo, err error) error {

		if !info.IsDir() {
			if info.Name() == ".DS_Store" || info.Name() == "Thumbs.db" {
				if err := os.Remove(path); err != nil {
					return err
				}
				fmt.Printf("  * deleted %s\n", path)
				deleteCount++
			}
		}

		return nil
	}); err != nil {
		return err
	}
	fmt.Printf("%d files deleted\n", deleteCount)
	return nil
}
