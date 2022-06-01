/*
Copyright 2022 Crunchy Data Solutions, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package bridgeapi

import "time"

type CreateRequest struct {
	Name             string `json:"name"`
	TeamID           string `json:"team_id"`
	Plan             string `json:"plan_id"`
	StorageGB        int    `json:"storage"`
	Provider         string `json:"provider_id"`
	Region           string `json:"region_id"`
	PGMajorVersion   int    `json:"postgres_version_id"`
	HighAvailability bool   `json:"is_ha"`
}

type ClusterList struct {
	Clusters []ClusterDetail `json:"clusters"`
}

type ClusterDetail struct {
	CPU              int       `json:"cpu"`
	Created          time.Time `json:"created_at"`
	ID               string    `json:"id"`
	HighAvailability bool      `json:"is_ha"`
	PGMajorVersion   int       `json:"major_version"`
	MaintWindowStart int       `json:"maintenance_window_start"`
	MemoryGB         int       `json:"memory"`
	Name             string    `json:"name"`
	PlanID           string    `json:"plan_id"`
	ProviderID       string    `json:"provider_id"`
	RegionID         string    `json:"region_id"`
	State            string    `json:"state"` // NOTE: Deprecated, but using to avoid extra status call on sync create for now
	StorageGB        int       `json:"storage"`
	TeamID           string    `json:"team_id"`
	Updated          time.Time `json:"updated_at"`
}

type ClusterStatus struct {
	DiskUsage      ClusterDiskUsage `json:"disk_usage"`
	OldestBackup   time.Time        `json:"oldest_backup_at"`
	OngoingUpgrade ClusterUpgrade   `json:"ongoing_upgrade"`
	State          string           `json:"state"`
}

type ClusterDiskUsage struct {
	Available int `json:"disk_available_mb"`
	Total     int `json:"disk_total_size_mb"`
	Used      int `json:"disk_used_mb"`
}

type ClusterUpgrade struct {
	Operations []ClusterUpgradeOperation
}

type ClusterUpgradeOperation struct {
	Flavor string `json:"flavor"`
	State  string `json:"state"`
}

type ClusterUpdateRequest struct {
	MaintWindowStart *int    `json:"maintenance_window_start,omitempty"`
	Name             *string `json:"name,omitempty"`
}

type ClusterUpgradeRequest struct {
	HighAvailability *bool   `json:"is_ha,omitempty"`
	PGMajorVersion   *int    `json:"postgres_version_id,omitempty"`
	PlanID           *string `json:"plan_id,omitempty"`
	StorageGB        *int    `json:"storage,omitempty"`
}

type ClusterRole struct {
	ClusterID string `json:"cluster_id"`
	Name      string `json:"name"`
	Password  string `json:"password"`
	TeamID    string `json:"team_id"`
	URI       string `json:"uri"`
}

type APIMessage struct {
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type Account struct {
	ID            string `json:"id"`
	DefaultTeamID string `json:"default_team_id"`
}

type Provider struct {
	ID       string       `json:"id"`
	Disk     ProviderDisk `json:"disk"`
	IconName string       `json:"icon_name"`
	Name     string       `json:"display_name"`
	Plans    []Plan       `json:"plans"`
	Regions  []Region     `json:"regions"`
}

type ProviderDisk struct {
	Rate int `json:"rate"`
}

type Plan struct {
	ID     string `json:"id"`
	CPU    int    `json:"cpu"`
	Memory int    `json:"memory"`
	Name   string `json:"display_name"`
	Rate   int    `json:"rate"`
}

type Region struct {
	ID         string  `json:"id"`
	Name       string  `json:"display_name"`
	Location   string  `json:"location"`
	Multiplier float64 `json:"multiplier"`
}

type Teams []Team

type Team struct {
	ID      string `json:"id"`
	Default bool   `json:"is_default"`
	Name    string `json:"name"`
	Role    string `json:"role"`
}
