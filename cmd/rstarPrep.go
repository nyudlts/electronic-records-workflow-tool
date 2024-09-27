package cmd

import (
	"fmt"
	"regexp"

	bagit "github.com/nyudlts/go-bagit"
)

var (
	woMatcher = regexp.MustCompile("aspace_wo.tsv$")
)

func prepPackage(bagLocation string, tmpLocation string) error {
	fmt.Println("ADOC Prep", version)

	bag, err := bagit.GetExistingBag(bagLocation)
	if err != nil {
		panic(err)
	}

	//validate the bag
	fmt.Printf("  * Validating bag at %s: ", bagLocation)
	if err := bag.ValidateBag(false, true); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	//move the work order to bag root and add to tag manifest
	fmt.Printf("  * Locating work order: ")
	matches := bag.Payload.FindFilesInPayload(woMatcher)
	if len(matches) != 0 {
		return fmt.Errorf("no workorder found")
	}
	woPath := matches[0].Path
	fmt.Printf("OK\n")

	fmt.Printf("  * Moving work order to bag's root ")
	if err := bag.AddFileToBagRoot(woPath); err != nil {
		return err
	}
	fmt.Printf("OK\n")

	return nil
}
