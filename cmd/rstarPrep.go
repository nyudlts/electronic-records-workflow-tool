package cmd

import (
	"fmt"

	bagit "github.com/nyudlts/go-bagit"
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

	return nil
}
