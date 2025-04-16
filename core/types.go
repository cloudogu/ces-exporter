package core

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type BackupSchedule struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
}

type GlobalConfig []KeyValue

type DoguConfig struct {
	Name            string     `json:"name"`
	NormalConfig    []KeyValue `json:"normal"`
	LocalConfig     []KeyValue `json:"local"`
	SensitiveConfig []KeyValue `json:"sensitive"`
}

type ExportResponse struct {
	GlobalConfig    GlobalConfig     `json:"global"`
	DoguConfigs     []DoguConfig     `json:"dogus"`
	BackupSchedules []BackupSchedule `json:"backupSchedules"`
}
