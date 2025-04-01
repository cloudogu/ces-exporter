package export

type DoguExport struct {
	Dogu         string `json:"dogu"`
	VolumePath   string `json:"volumePath"`
	ExporterPort int    `json:"exporterPort"`
}

type ExportModeStatus struct {
	IsActive bool `json:"isActive"`
}
