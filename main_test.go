package main

import (
	"bufio"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"
)

func TestFunctions(t *testing.T) {
	var transferInfo TransferInfo
	var workOrderComponent WorkOrderComponent
	t.Run("Test TransferInfo", func(t *testing.T) {
		transferInfo = TransferInfo{}
		transferInfoBytes, err := os.ReadFile("source/fales_mss267-electronic-records/metadata/transfer-info.txt")
		if err != nil {
			t.Error(err)
		}
		if err := yaml.Unmarshal(transferInfoBytes, &transferInfo); err != nil {
			t.Error(err)
		}

		t.Log(transferInfo)
	})
	t.Run("Create a WorkOrderComponent", func(t *testing.T) {
		workOrderFile, err := os.Open("source/fales_mss267-electronic-records/metadata/fales_mss267_aspace_wo.tsv")
		if err != nil {
			t.Error(err)
		}
		defer workOrderFile.Close()
		scanner := bufio.NewScanner(workOrderFile)
		scanner.Scan()
		scanner.Scan()
		line := strings.Split(scanner.Text(), "\t")
		workOrderComponent = WorkOrderComponent{line[0], line[1], line[2], line[3], line[4], line[5], line[6], line[7]}
	})

	t.Run("Create A DC MD file", func(t *testing.T) {
		dc := CreateDC(transferInfo, workOrderComponent)
		jBytes, err := json.MarshalIndent(&dc, "", "  ")
		if err != nil {
			t.Error(err)
		}

		t.Log(string(jBytes))
	})
}
