package cmd

import (
	"fmt"

	amatica "github.com/nyudlts/go-archivematica"
	"github.com/spf13/cobra"
)

var ingests bool
var transfers bool

func init() {
	clrCmd.Flags().StringVar(&amaticaConfigLoc, "config", "", "")
	clrCmd.Flags().BoolVar(&ingests, "ingests", false, "")
	clrCmd.Flags().BoolVar(&transfers, "transfers", false, "")
	amaticaCmd.AddCommand(clrCmd)
}

var clrCmd = &cobra.Command{
	Use: "clear",
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkFlags(); err != nil {
			panic(err)
		}

		if err := clear(); err != nil {
			panic(err)
		}
	},
}

func clear() error {

	var err error
	client, err = amatica.NewAMClient(amaticaConfigLoc, 20)
	if err != nil {
		return err
	}

	if transfers {
		fmt.Println("Clearing completed transfers")
		completedTransfers, err := client.GetCompletedTransfers()
		if err != nil {
			return err
		}

		completedTransfersMap, err := client.GetCompletedTransfersMap(completedTransfers)
		if err != nil {
			return err
		}

		for k, v := range completedTransfersMap {
			fmt.Printf("clearing %s: %s\n", k, v.Name)
			if err := client.DeleteTransfer(v.UUID); err != nil {
				return err
			}
			fmt.Printf("%s: %s cleared\n", k, v.Name)
		}
	}

	if ingests {
		completedIngests, err := client.GetCompletedIngests()
		if err != nil {
			return err
		}

		completedIngestsMap, err := client.GetCompletedIngestsMap(completedIngests)
		if err != nil {
			return err
		}

		for k, v := range completedIngestsMap {
			fmt.Printf("clearing %s: %s\n", k, v.Name)
			if err := client.DeleteIngest(v.UUID); err != nil {
				return err
			}
			fmt.Printf("%s: %s cleared\n", k, v.Name)
		}
	}

	return nil
}
