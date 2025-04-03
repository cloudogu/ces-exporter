package maintenance

type maintenanceModeRequest struct {
	Activate bool    `json:"activate"`
	Message  Message `json:"message"`
}

type Message struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type MaintenanceModeStatus struct {
	IsActive bool `json:"isActive"`
}
