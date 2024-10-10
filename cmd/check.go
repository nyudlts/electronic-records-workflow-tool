package cmd

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
)

func init() {
	checkCmd.Flags().StringVar(&aspaceConfigLoc, "aspace-config", "", "")
	checkCmd.Flags().StringVar(&aspaceWOLoc, "aspace-workorder", "", "")
	checkCmd.Flags().StringVar(&aspaceEnv, "aspace-environment", "", "")
	rootCmd.AddCommand(checkCmd)
}

var checkCmd = &cobra.Command{
	Use: "check",
	Run: func(cmd *cobra.Command, args []string) {
		aspaceCheck()
	},
}

func aspaceCheck() {
	client, err := aspace.NewClient(aspaceConfigLoc, aspaceEnv, 20)
	if err != nil {
		panic(err)
	}

	workOrder, _ := os.Open(aspaceWOLoc)
	defer workOrder.Close()
	wo := aspace.WorkOrder{}
	if err := wo.Load(workOrder); err != nil {
		panic(err)
	}

	var b bytes.Buffer
	out := csv.NewWriter(bufio.NewWriter(&b))
	out.Comma = '\t'
	out.Write([]string{"ao_uri", "title", "do_uri", "do_id", "msg"})

	for _, row := range wo.Rows {
		repoId, aoURI, err := aspace.URISplit(row.GetURI())
		if err != nil {
			panic(err)
		}

		ao, err := client.GetArchivalObject(repoId, aoURI)
		if err != nil {
			panic(err)
		}

		instances := ao.Instances

		if len(instances) < 1 {
			out.Write([]string{ao.URI, ao.Title, "ERROR: AO has no instances", ao.ComponentId, "KO"})
			out.Flush()
			continue
		}

		for _, instance := range instances {
			if instance.InstanceType == "digital_object" {
				doURI := instance.DigitalObject["ref"]
				_, doID, err := aspace.URISplit(doURI)
				if err != nil {
					out.Write([]string{row.GetURI(), "", "", "ERROR: Not able to split: " + doURI})
					out.Flush()
					continue
				}

				do, err := client.GetDigitalObject(repoId, doID)
				if err != nil {
					out.Write([]string{row.GetURI(), "", "", "ERROR: not able to request: " + doURI})
					out.Flush()
					continue
				}

				if do.DigitalObjectID != row.GetComponentID() {
					fmt.Printf("DO: `%s` WO: `%s`", do.DigitalObjectID, row.GetComponentID())
					out.Write([]string{row.GetURI(), do.URI, do.DigitalObjectID, "ERROR: component IDs do not match"})
					out.Flush()
					continue
				} else {
					out.Write([]string{row.GetURI(), do.Title, do.URI, do.DigitalObjectID, "OK"})
					out.Flush()
					continue
				}

			}
		}
	}

	if err := os.WriteFile("adoc-check.tsv", b.Bytes(), 0777); err != nil {
		panic(err)
	}

}
