// Copyright 2019 Yunion
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compute

import (
	"yunion.io/x/jsonutils"

	"yunion.io/x/onecloud/pkg/apis"
)

type HostSpec struct {
	apis.Meta

	Cpu             int                  `json:"cpu"`
	Mem             int                  `json:"mem"`
	NicCount        int                  `json:"nic_count"`
	Manufacture     string               `json:"manufacture"`
	Model           string               `json:"model"`
	Disk            DiskDriverSpec       `json:"disk"`
	Driver          string               `json:"driver"`
	IsolatedDevices []IsolatedDeviceSpec `json:"isolated_devices"`
}

type IsolatedDeviceSpec struct {
	apis.Meta

	DevType string `json:"dev_type"`
	Model   string `json:"model"`
	PciId   string `json:"pci_id"`
	Vendor  string `json:"vendor"`
}

type DiskDriverSpec map[string]DiskAdapterSpec

type DiskAdapterSpec map[string][]*DiskSpec

type DiskSpec struct {
	apis.Meta

	Type       string `json:"type"`
	Size       int64  `json:"size"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
	Count      int    `json:"count"`
}

type HostListInput struct {
	apis.EnabledStatusInfrasResourceBaseListInput
	apis.ExternalizedResourceBaseListInput

	ManagedResourceListInput
	ZonalFilterListInput
	WireFilterListInput
	SchedtagFilterListInput

	StorageFilterListInput
	UsableResourceListInput

	// filter by ResourceType
	ResourceType string `json:"resource_type"`
	// filter by mac of any network interface
	AnyMac string `json:"any_mac"`
	// filter by ip of any network interface
	AnyIp string `json:"any_ip"`
	// filter storages not attached to this host
	StorageNotAttached *bool `json:"storage_not_attached"`
	// filter by Hypervisor
	Hypervisor string `json:"hypervisor"`
	// filter host that is empty
	IsEmpty *bool `json:"is_empty"`
	// filter host that is baremetal
	Baremetal *bool `json:"baremetal"`

	// ??????
	Rack []string `json:"rack"`
	// ??????
	Slots []string `json:"slots"`
	// ?????????MAC
	AccessMac []string `json:"access_mac"`
	// ?????????Ip??????
	AccessIp []string `json:"access_ip"`
	// ????????????????????????
	SN []string `json:"sn"`
	// CPU??????
	CpuCount []int `json:"cpu_count"`
	// ????????????,??????Mb
	MemSize []int `json:"mem_size"`
	// ????????????
	StorageType []string `json:"storage_type"`
	// IPMI??????
	IpmiIp []string `json:"ipmi_ip"`
	// ???????????????
	// example: online
	HostStatus []string `json:"host_status"`
	// ???????????????
	HostType []string `json:"host_type"`
	// host??????????????????
	Version []string `json:"version"`
	// OVN????????????
	OvnVersion []string `json:"ovn_version"`
	// ????????????????????????
	IsMaintenance *bool `json:"is_maintenance"`
	// ???????????????????????????
	IsImport *bool `json:"is_import"`
	// ????????????PXE??????
	EnablePxeBoot *bool `json:"enable_pxe_boot"`
	// ??????UUID
	Uuid []string `json:"uuid"`
	// ??????????????????, ????????????PXE???ISO
	BootMode []string `json:"boot_mode"`
	// ??????????????????????????????
	ServerIdForNetwork string `json:"server_id_for_network"`
	// ????????? cpu ??????
	CpuArchitecture []string `json:"cpu_architecture"`
	OsArch          string   `json:"os_arch"`

	// ????????????????????????
	// enum: asc,desc
	OrderByServerCount string `json:"order_by_server_count"`
	// ?????????????????????
	// enmu: asc,desc
	OrderByStorage string `json:"order_by_storage"`
}

type HostDetails struct {
	apis.EnabledStatusInfrasResourceBaseDetails
	ManagedResourceInfo
	ZoneResourceInfo

	SHost

	Schedtags []SchedtagShortDescDetails `json:"schedtags"`

	ServerId             string `json:"server_id"`
	Server               string `json:"server"`
	ServerIps            string `json:"server_ips"`
	ServerPendingDeleted bool   `json:"server_pending_deleted"`
	// ????????????
	NicCount int `json:"nic_count"`
	// ????????????
	NicInfo []jsonutils.JSONObject `json:"nic_info"`
	// CPU?????????
	CpuCommit int `json:"cpu_commit"`
	// ???????????????
	MemCommit int `json:"mem_commit"`
	// ???????????????
	// example: 10
	Guests int `json:"guests,allowempty"`
	// ????????????????????????
	// example: 0
	NonsystemGuests int `json:"nonsystem_guests,allowempty"`
	// ???????????????????????????
	// example: 2
	RunningGuests int `json:"running_guests,allowempty"`
	// ???????????????????????????
	// example: 2
	ReadyGuests int `json:"ready_guests,allowempty"`
	// ???????????????????????????
	// example: 2
	OtherGuests int `json:"other_guests,allowempty"`
	// ???????????????????????????
	// example: 2
	PendingDeletedGuests int `json:"pending_deleted_guests,allowempty"`
	// CPU?????????
	CpuCommitRate float64 `json:"cpu_commit_rate"`
	// ???????????????
	MemCommitRate float64 `json:"mem_commit_rate"`
	// CPU?????????
	CpuCommitBound float32 `json:"cpu_commit_bound"`
	// ???????????????
	MemCommitBound float32 `json:"mem_commint_bound"`
	// ????????????
	Storage int64 `json:"storage"`
	// ?????????????????????
	StorageUsed int64 `json:"storage_used"`
	// ???????????????????????????
	ActualStorageUsed int64 `json:"actual_storage_used"`
	// ??????????????????(????????????????????????)
	StorageWaste int64 `json:"storage_waste"`
	// ??????????????????
	StorageVirtual int64 `json:"storage_virtual"`
	// ??????????????????
	StorageFree int64 `json:"storage_free"`
	// ???????????????
	StorageCommitRate float64 `json:"storage_commit_rate"`

	Spec              *jsonutils.JSONDict `json:"spec"`
	IsPrepaidRecycle  bool                `json:"is_prepaid_recycle"`
	CanPrepare        bool                `json:"can_prepare"`
	PrepareFailReason string              `json:"prepare_fail_reason"`
	// ?????????????????????????????????
	AllowHealthCheck      bool `json:"allow_health_check"`
	AutoMigrateOnHostDown bool `json:"auto_migrate_on_host_down"`

	// reserved resource for isolated device
	ReservedResourceForGpu IsolatedDeviceReservedResourceInput `json:"reserved_resource_for_gpu"`
	// isolated device count
	IsolatedDeviceCount int

	// host init warnning
	SysWarn string `json:"sys_warn"`
	// host init error info
	SysError string `json:"sys_error"`
}

type HostResourceInfo struct {
	// ???????????????ID
	ManagerId string `json:"manager_id"`

	ManagedResourceInfo

	// ???????????????ID
	ZoneId string `json:"zone_id"`

	ZoneResourceInfo

	// ???????????????
	Host string `json:"host"`

	// ??????????????????
	HostSN string `json:"host_sn"`

	// ???????????????
	HostStatus string `json:"host_status"`

	// ?????????????????????`
	HostServiceStatus string `json:"host_service_status"`

	// ???????????????
	HostType string `json:"host_type"`
}

type HostFilterListInput struct {
	ZonalFilterListInput
	ManagedResourceListInput

	HostFilterListInputBase
}

type HostFilterListInputBase struct {
	HostResourceInput

	// ???????????????????????????
	HostSN string `json:"host_sn"`

	// ????????????????????????
	OrderByHost string `json:"order_by_host"`

	// ?????????????????????????????????
	OrderByHostSN string `json:"order_by_host_sn"`
}

type HostResourceInput struct {
	// ????????????????????????ID???Name???
	HostId string `json:"host_id"`
	// swagger:ignore
	// Deprecated
	// filter by host_id
	Host string `json:"host" yunion-deprecated-by:"host_id"`
}

type HostRegisterMetadata struct {
	apis.Meta

	OnKubernetes                 bool   `json:"on_kubernetes"`
	Hostname                     string `json:"hostname"`
	SysError                     string `json:"sys_error,allowempty"`
	SysWarn                      string `json:"sys_warn,allowempty"`
	RootPartitionTotalCapacityMB int64  `json:"root_partition_total_capacity_mb"`
	RootPartitionUsedCapacityMB  int64  `json:"root_partition_used_capacity_mb"`
}

type HostAccessAttributes struct {
	// ???????????????URI
	ManagerUri string `json:"manager_uri"`

	// ??????????????????IP
	AccessIp string `json:"access_ip"`

	// ??????????????????MAC
	AccessMac string `json:"access_mac"`

	// ??????????????????IP??????
	AccessNet string `json:"access_net"`
	// ??????????????????????????????
	AccessWire string `json:"access_wire"`
}

type HostSizeAttributes struct {
	// CPU??????
	CpuCount *int `json:"cpu_count"`
	// ??????CPU??????
	NodeCount *int8 `json:"node_count"`
	// CPU????????????
	CpuDesc string `json:"cpu_desc"`
	// CPU??????
	CpuMhz *int `json:"cpu_mhz"`
	// CPU????????????,??????KB
	CpuCache string `json:"cpu_cache"`
	// ??????CPU??????
	CpuReserved *int `json:"cpu_reserved"`
	// CPU?????????
	CpuCmtbound *float32 `json:"cpu_cmtbound"`
	// CPUMicrocode
	CpuMicrocode string `json:"cpu_microcode"`
	// CPU??????
	CpuArchitecture string `json:"cpu_architecture"`

	// ????????????(??????MB)
	MemSize string `json:"mem_size"`
	// ??????????????????(??????MB)
	MemReserved string `json:"mem_reserved"`
	// ???????????????
	MemCmtbound *float32 `json:"mem_cmtbound"`

	// ????????????,??????Mb
	StorageSize *int `json:"storage_size"`
	// ????????????
	StorageType string `json:"storage_type"`
	// ??????????????????
	StorageDriver string `json:"storage_driver"`
	// ????????????
	StorageInfo jsonutils.JSONObject `json:"storage_info"`
}

type HostIpmiAttributes struct {
	// username
	IpmiUsername string `json:"ipmi_username"`
	// password
	IpmiPassword string `json:"ipmi_password"`
	// ip address
	IpmiIpAddr string `json:"ipmi_ip_addr"`
	// presence
	IpmiPresent *bool `json:"ipmi_present"`
	// lan channel
	IpmiLanChannel *int `json:"ipmi_lan_channel"`
	// verified
	IpmiVerified *bool `json:"ipmi_verified"`
	// Redfish API support
	IpmiRedfishApi *bool `json:"ipmi_redfish_api"`
	// Cdrom boot support
	IpmiCdromBoot *bool `json:"ipmi_cdrom_boot"`
	// ipmi_pxe_boot
	IpmiPxeBoot *bool `json:"ipmi_pxe_boot"`
}

type HostCreateInput struct {
	apis.EnabledStatusInfrasResourceBaseCreateInput

	ZoneResourceInput
	HostnameInput

	HostAccessAttributes
	HostSizeAttributes
	HostIpmiAttributes

	// ?????????IPMI??????????????????????????????IPMI????????????
	NoProbe *bool `json:"no_probe"`

	// host uuid
	Uuid string `json:"uuid"`

	// Host??????
	HostType string `json:"host_type"`

	// ??????????????????
	IsBaremetal *bool `json:"is_baremetal"`

	// ??????
	Rack string `json:"rack"`
	// ??????
	Slots string `json:"slots"`

	// ????????????
	SysInfo jsonutils.JSONObject `json:"sys_info"`

	// ????????????????????????
	SN string `json:"sn"`

	// host??????????????????
	Version string `json:"version"`
	// OVN????????????
	OvnVersion string `json:"ovn_version"`

	// ???????????????????????????
	IsImport *bool `json:"is_import"`

	// ????????????PXE??????
	EnablePxeBoot *bool `json:"enable_pxe_boot"`

	// ??????????????????, ????????????PXE???ISO
	BootMode string `json:"boot_mode"`
}

type HostUpdateInput struct {
	apis.EnabledStatusInfrasResourceBaseUpdateInput

	HostAccessAttributes
	HostSizeAttributes
	HostIpmiAttributes

	// IPMI info
	IpmiInfo jsonutils.JSONObject `json:"ipmi_info"`

	// ??????
	Rack string `json:"rack"`
	// ??????
	Slots string `json:"slots"`

	// ????????????
	SysInfo jsonutils.JSONObject `json:"sys_info"`
	// ????????????????????????
	SN string `json:"sn"`

	// ???????????????
	HostType string `json:"host_type"`

	// host??????????????????
	Version string `json:"version"`
	// OVN????????????
	OvnVersion string `json:"ovn_version"`
	// ??????????????????
	IsBaremetal *bool `json:"is_baremetal"`

	// ????????????PXE??????
	EnablePxeBoot *bool `json:"enable_pxe_boot"`

	// ??????UUID
	Uuid string `json:"uuid"`

	// ??????????????????, ????????????PXE???ISO
	BootMode string `json:"boot_mode"`
}

type HostOfflineInput struct {
	UpdateHealthStatus *bool `json:"update_health_status"`
	Reason             string
}
