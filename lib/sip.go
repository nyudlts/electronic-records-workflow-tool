package lib

import (
	"io/fs"
	"os"
	"path/filepath"
)

func CleanSip() error {
	//load the project configuration
	if err := loadConfig(); err != nil {
		return err
	}

	if err := filepath.Walk(config.SIPLoc, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			if info.Name() == ".DS_Store" || info.Name() == "Thumbs.db" {
				if err := os.Remove(path); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
