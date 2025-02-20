package cmd

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/nyudlts/go-aspace"
	"github.com/spf13/cobra"
)

var aspaceEnv string
var aspaceConfigLoc string

func init() {
	checkCmd.Flags().StringVar(&aspaceConfigLoc, "aspace-config", "", "if not set will default to `/home/'username'/.config/go-aspace.yml")
	checkCmd.Flags().StringVar(&aspaceEnv, "aspace-environment", "prod", "the environment to to lookup in config")
	aspaceCmd.AddCommand(checkCmd)
}

var workOrderLocation string

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check that DOs exist in Archivesspace",
	Run: func(cmd *cobra.Command, args []string) {
		//print bin vers and cmd
		fmt.Printf("ADOC %s ASPACE CHECK\n", version)

		//load project config
		if err := loadProjectConfig(); err != nil {
			panic(err)
		}

		//get aspaceConfig
		if err := getConfig(); err != nil {
			panic(err)
		}

		//get workorder
		if err := findWorkOrder(); err != nil {
			panic(err)
		}

		//run the check
		if err := aspaceCheck(); err != nil {
			panic(err)
		}
	},
}

func getConfig() error {
	if aspaceConfigLoc == "" {
		currentUser, err := user.Current()
		if err != nil {
			return (err)
		}
		aspaceConfigLoc = fmt.Sprintf("/home/%s/.config/go-aspace.yml", currentUser.Username)
	}

	_, err := os.Stat(aspaceConfigLoc)
	if err != nil {
		return err
	}
	return nil
}

func findWorkOrder() error {
	mdDir := filepath.Join(adocConfig.StagingLoc, "metadata")
	var err error
	workOrderFilename, err := getWorkOrderFile(mdDir)
	if err != nil {
		return err
	}
	workOrderLocation = filepath.Join(mdDir, workOrderFilename)
	return nil
}

func aspaceCheck() error {
	client, err := aspace.NewClient(aspaceConfigLoc, aspaceEnv, 20)
	if err != nil {
		panic(err)
	}

	workOrder, _ := os.Open(workOrderLocation)
	defer workOrder.Close()
	wo := aspace.WorkOrder{}
	if err := wo.Load(workOrder); err != nil {
		panic(err)
	}

	var b bytes.Buffer
	out := csv.NewWriter(bufio.NewWriter(&b))
	out.Comma = '\t'
	out.Write([]string{"ao_uri", "title", "do_uri", "do_id", "msg"})
	out.Flush()

	for _, row := range wo.Rows {
		repoId, aoURI, err := aspace.URISplit(row.GetURI())
		if err != nil {
			return err
		}
		fmt.Printf("Checking: %s\n", row.GetURI())

		ao, err := client.GetArchivalObject(repoId, aoURI)
		if err != nil {
			out.Write([]string{row.GetURI(), "", "", "ERROR: AO does not exist: " + row.GetURI()})
			out.Flush()
			continue
		}
		fmt.Println("Found AO:", ao.URI)

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

					out.Write([]string{row.GetURI(), do.URI, do.DigitalObjectID, "ERROR: component IDs do not match"})
					fmt.Printf("Component IDs do not match: %s, %s, %s\n", row.GetURI(), do.URI, do.DigitalObjectID)
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

	checkFilename := filepath.Join("logs", fmt.Sprintf("%s-adoc-check.tsv", adocConfig.CollectionCode))

	if err := os.WriteFile(checkFilename, b.Bytes(), 0775); err != nil {
		panic(err)
	}

	fmt.Println("Checkfile written to:", checkFilename)

	return nil

}
