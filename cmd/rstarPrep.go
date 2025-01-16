package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	bagit "github.com/nyudlts/go-bagit"
)

var (
	woMatcher = regexp.MustCompile("aspace_wo.tsv$")
	tiMatcher = regexp.MustCompile("transfer-info.txt")
)

func prepPackage(bagLocation string, tmpLocation string) error {

	bag, err := bagit.GetExistingBag(bagLocation)
	if err != nil {
		return err
	}

	//validate the bag
	fmt.Printf("  * Validating bag at %s: ", bagLocation)
	if err := bag.ValidateBag(false, false); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//Locate the workorder
	fmt.Printf("  * Locating work order: ")
	matches := bag.Payload.FindFilesInPayload(woMatcher)
	if len(matches) != 1 {
		return fmt.Errorf("no workorder found")
	}
	woPath := matches[0].Path
	fmt.Printf("OK\n")

	//Move the workorder to the root of the bag
	fmt.Printf("  * Moving work order to bag's root: ")
	if err := bag.AddFileToBagRoot(woPath); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//Locate transfer-info.txt
	fmt.Printf("  * Locating transfer-info.txt: ")
	matches = bag.Payload.FindFilesInPayload(tiMatcher)
	if len(matches) != 1 {
		return fmt.Errorf("no transfer-info.txt found")
	}
	tiPath := matches[0].Path
	tiPath = strings.ReplaceAll(tiPath+"/", bagLocation, "")
	fmt.Printf("OK\n")

	//create a tag set from transfer-info.txt
	fmt.Printf("  * Creating new tag set from %s: ", "transfer-info.txt")
	transferInfo, err := bagit.NewTagSet(tiPath, bagLocation)
	if err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//Update the hostname
	fmt.Printf("  * Adding hostname to tag set: ")
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	transferInfo.Tags["nyu-dl-hostname"] = hostname
	fmt.Printf("OK\n")

	//add pathname to the tag-set
	fmt.Printf("  * Adding bag's path to tag set: ")
	path, err := filepath.Abs(bagLocation)
	if err != nil {
		return err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	transferInfo.Tags["nyu-dl-pathname"] = path
	fmt.Printf("OK\n")

	//backup bag-info
	fmt.Print("  * Backing up bag-info.txt")
	bagInfoLocation := filepath.Join(bagLocation, "bag-info.txt")
	backupLocation := filepath.Join(tmpLocation, "bag-info.txt")
	backup, err := os.Create(backupLocation)
	if err != nil {
		return err
	}
	defer backup.Close()

	source, err := os.Open(bagInfoLocation)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = io.Copy(backup, source)
	if err != nil {
		return err
	}
	fmt.Printf(" OK\n")

	//getting tagset from bag-info
	fmt.Printf("  * Creating new tag set from %s: ", "bag-info.txt")
	bagInfo, err := bagit.NewTagSet("bag-info.txt", bagLocation)
	if err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//merge tagsets
	fmt.Printf("  * Merging Tag Sets: ")
	bagInfo.AddTags(transferInfo.Tags)
	fmt.Printf("OK\n")

	fmt.Printf("  * Getting data as byte array: ")
	bagInfoBytes := bagInfo.GetTagSetAsByteSlice()
	fmt.Printf("OK\n")

	fmt.Printf("  * Opening bag-info.txt: ")
	bagInfoFile, err := os.Open(bagInfoLocation)
	if err != nil {
		return err
	}
	defer bagInfoFile.Close()
	fmt.Printf("OK\n")

	fmt.Printf("  * Rewriting bag-info.txt: ")
	if err := os.WriteFile(bagInfoLocation, bagInfoBytes, 0777); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//create new manifest object for tagmanifest-sha256.txt
	fmt.Printf("  * Creating new tagmanifest-sha256.txt: ")
	tagManifest, err := bagit.NewManifest(bagLocation, "tagmanifest-sha256.txt")
	if err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//update the checksum for bag-info.txt
	fmt.Printf("  * Updating checksum for bag-info.txt in tagmanifest-sha256.txt: ")
	if err := tagManifest.UpdateManifest("bag-info.txt"); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	fmt.Printf("  * Rewriting tagmanifest-sha256.txt: ")
	if err := tagManifest.Serialize(); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//validate the updated bag
	fmt.Printf("\nValidating the updated bag: ")
	if err := bag.ValidateBag(false, false); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//delete the backup bag-info
	fmt.Printf("  * Deleting backup bag-info.txt: ")
	if err := os.Remove(backupLocation); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	fmt.Println("\nPackage preparation complete")

	return nil
}
