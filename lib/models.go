package lib

type Config struct {
	SIPLoc           string `yaml:"sip-location"`
	SourceLoc        string `yaml:"source-location"`
	PartnerCode      string `yaml:"partner-code"`
	CollectionCode   string `yaml:"collection-code"`
	ProjectLoc       string `yaml:"project-location"`
	LogLoc           string `yaml:"log-location"`
	AIPLoc           string `yaml:"aip-location"`
	AMTransferSource string `yaml:"archivematica-transfer-source"`
	XferLoc          string `yaml:"xfer-location"`
}
