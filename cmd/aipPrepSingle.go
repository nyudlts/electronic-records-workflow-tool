package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	prepSingleCmd.Flags().StringVar(&aipLoc, "aip-location", "", "location of aip to be prepped (required)")
	prepSingleCmd.Flags().StringVar(&tmpLoc, "tmp-location", ".", "location for temp processing")
	aipCmd.AddCommand(prepSingleCmd)
}

var prepSingleCmd = &cobra.Command{
	Use:   "prep-single",
	Short: "Prepare a single AIP for transfer to R*",
	Run: func(cmd *cobra.Command, args []string) {
		prepSingle()
	},
}

func prepSingle() {
	fmt.Println("ADOC aip prep", version)
	fmt.Printf("Prepping bag at %s for transfer to R*\n", aipLoc)
	//check that aip exists
	aip, err := os.Stat(aipLoc)
	if err != nil {
		panic(err)
	}

	if !aip.IsDir() {
		panic(fmt.Errorf("%s is not a directory", aip))
	}

	tmp, err := os.Stat(tmpLoc)
	if err != nil {
		panic(err)
	}

	if !tmp.IsDir() {
		panic(fmt.Errorf("%s is not a directory", tmp))
	}

	if err := prepPackage(aipLoc, tmpLoc); err != nil {
		panic(err)
	}

}
