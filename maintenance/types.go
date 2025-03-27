package maintenance

type maintenanceModeRequest struct {
	Activate bool   `json:"activate"`
	Title    string `json:"title"`
	Message  string `json:"message"`
}

type MaintenanceModeStatus struct {
	IsActive bool `json:"isActive"`
}
