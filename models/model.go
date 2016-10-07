package model

import (
	"time"
)

type BackupStatus string

const (
	CreateSucceeded BackupStatus = "CREATE_SUCCEEDED"
	CreateInProgress BackupStatus = "CREATE_IN_PROGRESS"
	DeleteInProgress BackupStatus = "DELETE_IN_PROGRESS"
)

type RestoreStatus string

const (
	Succeeded RestoreStatus = "SUCCEEDED"
)

type RestoreEntity struct  {
	BackupId string `json:"backup_id"`
	Status RestoreStatus `json:"status"`
}

type BackupRestore struct  {
	Metadata Metadata `json:"metadata"`
	Entity RestoreEntity `json:"entity"`
}

type Metadata struct {
	GUID string `json:"guid"`
	URL string `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BackupEntity struct {
	ServiceInstanceId string `json:"service_instance_id"`
	Status BackupStatus `json:"status"`
	Restores []BackupRestore `json:"restores"`
}

type ServiceInstanceBackup struct {
	Metadata *Metadata `json:"metadata"`
	Entity *BackupEntity `json:"entity"`
}

type ServiceInstanceResults struct {
	TotalResults int `json:"total_results"`
	TotalPages int `json:"total_pages"`
	PrevURL *string `json:"prev_url"`
	NextURL *string `json:"next_url"`
	Resources []ServiceInstanceBackup `json:"resources"`
}