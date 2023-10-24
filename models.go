package main

import (
	"fmt"
	"strconv"
	"strings"
)

// {"title": "Electronic Records", "is_part_of": "AIC#MSS.267: Alan Sondheim Papers"}
type DC struct {
	Title    string `json:"title"`
	IsPartOf string `json:"is_part_of"`
}

type WorkOrderComponent struct {
	ResourceID          string
	RefID               string
	URI                 string
	ContainerIndicator1 string
	ContainerIndicator2 string
	ContainerIndicator3 string
	Title               string
	ComponentID         string
}

func (w WorkOrderComponent) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", w.ResourceID, w.RefID, w.URI, w.ContainerIndicator1, w.ContainerIndicator2, w.ContainerIndicator3, w.Title, w.ComponentID)
}

func (w WorkOrderComponent) GetERID() (int, error) {
	split := strings.Split(w.ContainerIndicator2, "_")
	return strconv.Atoi(split[4])
}

type TransferInfo struct {
	ContactName              string `yaml:"Contact-Name"`
	ContactPhone             string `yaml:"Contact-Phone"`
	ContactEmail             string `yaml:"Contact-Email"`
	InternalSenderIdentifier string `yaml:"Internal-Sender-Identifier"`
	OrganizationAdress       string `yaml:"OrganizationAddress"`
	SourceOrganization       string `yaml:"SourceOrganization"`
	ArchivesSpaceResourceURL string `yaml:"nyu-dl-archivesspace-resource-url"`
	ResourceID               string `yaml:"nyu-dl-resource-id"`
	ResourceTitle            string `yaml:"nyu-dl-resource-title"`
	ContrentType             string `yaml:"nyu-dl-content-type"`
	ContentClassification    string `yaml:"nyu-dl-content-classification"`
	ProjectName              string `yaml:"nyu-dl-project-name"`
	RStarCollectionID        string `yaml:"nyu-dl-rstar-collection-id"`
}
