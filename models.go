package main

// {"title": "Electronic Records", "is_part_of": "AIC#MSS.267: Alan Sondheim Papers"}
type DC struct {
	Title    string `json:"title"`
	IsPartOf string `json:"is_part_of"`
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
