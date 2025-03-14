package export

type doguExport struct {
	Dogu         string `json:"dogu"`
	VolumePath   string `json:"volumePath"`
	ExporterPort int    `json:"exporterPort"`
}

type exportModeStatus struct {
	IsActive bool `json:"isActive"`
}
