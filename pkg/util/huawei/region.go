package huawei

import (
	"fmt"
	"strconv"

	"strings"
	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/util/huawei/client"
	"yunion.io/x/pkg/util/secrules"
)

type Locales struct {
	EnUs string `json:"en-us"`
	ZhCN string `json:"zh-cn"`
}

// https://support.huaweicloud.com/api-iam/zh-cn_topic_0067148043.html
type SRegion struct {
	client    *SHuaweiClient
	ecsClient *clients.Client

	Description    string  `json:"description"`
	ID             string  `json:"id"`
	Locales        Locales `json:"locales"`
	ParentRegionID string  `json:"parent_region_id"`
	Type           string  `json:"type"`

	izones []cloudprovider.ICloudZone
	ivpcs  []cloudprovider.ICloudVpc

	storageCache *SStoragecache
}

func (self *SRegion) getECSClient() (*clients.Client, error) {
	var err error

	if len(self.client.projectId) > 0 {
		project, err := self.client.GetProjectById(self.client.projectId)
		if err != nil {
			return nil, err
		}

		regionId := strings.Split(project.Name, "_")[0]
		if regionId != self.ID {
			log.Debugf("project %s not in region %s", self.client.projectId, self.ID)
			return nil, nil
		}
	}

	if self.ecsClient == nil {
		self.ecsClient, err = clients.NewClientWithAccessKey(self.ID, self.client.projectId, self.client.accessKey, self.client.secret)
		if err != nil {
			return nil, err
		}

		return self.ecsClient, err
	}

	return self.ecsClient, err
}

func (self *SRegion) fetchZones() error {
	zones := make([]SZone, 0)
	err := DoList(self.ecsClient.Zones.List, nil, &zones)
	if err != nil {
		return err
	}

	self.izones = make([]cloudprovider.ICloudZone, 0)
	for i := range zones {
		zone := zones[i]
		zone.region = self
		self.izones = append(self.izones, &zone)
	}
	return nil
}

func (self *SRegion) fetchIVpcs() error {
	// https://support.huaweicloud.com/api-vpc/zh-cn_topic_0020090625.html
	vpcs := make([]SVpc, 0)
	querys := map[string]string{
		"limit": "2048",
	}
	err := DoList(self.ecsClient.Vpcs.List, querys, &vpcs)
	if err != nil {
		return err
	}

	self.ivpcs = make([]cloudprovider.ICloudVpc, 0)
	for i := range vpcs {
		vpc := vpcs[i]
		vpc.region = self
		self.ivpcs = append(self.ivpcs, &vpc)
	}
	return nil
}

func (self *SRegion) GetGeographicInfo() cloudprovider.SGeographicInfo {
	if info, ok := LatitudeAndLongitude[self.ID]; ok {
		return info
	}
	return cloudprovider.SGeographicInfo{}
}

func (self *SRegion) GetILoadBalancers() ([]cloudprovider.ICloudLoadbalancer, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (self *SRegion) GetILoadbalancerAcls() ([]cloudprovider.ICloudLoadbalancerAcl, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (self *SRegion) GetILoadbalancerCertificates() ([]cloudprovider.ICloudLoadbalancerCertificate, error) {
	return nil, cloudprovider.ErrNotImplemented
}

// https://support.huaweicloud.com/api-iam/zh-cn_topic_0057845622.html
func (self *SRegion) GetId() string {
	return self.ID
}

func (self *SRegion) GetName() string {
	return fmt.Sprintf("%s %s", CLOUD_PROVIDER_HUAWEI_CN, self.Locales.ZhCN)
}

func (self *SRegion) GetGlobalId() string {
	return fmt.Sprintf("%s/%s", CLOUD_PROVIDER_HUAWEI, self.ID)
}

func (self *SRegion) GetStatus() string {
	return models.CLOUD_REGION_STATUS_INSERVER
}

func (self *SRegion) Refresh() error {
	return nil
}

func (self *SRegion) IsEmulated() bool {
	return false
}

func (self *SRegion) GetMetadata() *jsonutils.JSONDict {
	return nil
}

func (self *SRegion) GetLatitude() float32 {
	if locationInfo, ok := LatitudeAndLongitude[self.ID]; ok {
		return locationInfo.Latitude
	}
	return 0.0
}

func (self *SRegion) GetLongitude() float32 {
	if locationInfo, ok := LatitudeAndLongitude[self.ID]; ok {
		return locationInfo.Longitude
	}
	return 0.0
}

func (self *SRegion) fetchInfrastructure() error {
	_, err := self.getECSClient()
	if err != nil {
		return err
	}

	if err := self.fetchZones(); err != nil {
		return err
	}

	if err := self.fetchIVpcs(); err != nil {
		return err
	}

	for i := 0; i < len(self.ivpcs); i += 1 {
		for j := 0; j < len(self.izones); j += 1 {
			zone := self.izones[j].(*SZone)
			vpc := self.ivpcs[i].(*SVpc)
			wire := SWire{zone: zone, vpc: vpc}
			zone.addWire(&wire)
			vpc.addWire(&wire)
		}
	}
	return nil
}

func (self *SRegion) GetIZones() ([]cloudprovider.ICloudZone, error) {
	if self.izones == nil {
		var err error
		err = self.fetchInfrastructure()
		if err != nil {
			return nil, err
		}
	}
	return self.izones, nil
}

func (self *SRegion) GetIVpcs() ([]cloudprovider.ICloudVpc, error) {
	if self.ivpcs == nil {
		err := self.fetchInfrastructure()
		if err != nil {
			return nil, err
		}
	}
	return self.ivpcs, nil
}

func (self *SRegion) GetEipById(eipId string) (SEipAddress, error) {
	var eip SEipAddress
	err := DoGet(self.ecsClient.Eips.Get, eipId, nil, &eip)
	eip.region = self
	return eip, err
}

// 返回参数分别为eip 列表、列表长度、error。
func (self *SRegion) GetEips(marker string, limit int) ([]SEipAddress, int, error) {
	querys := map[string]string{"limit": "50"}
	if len(marker) > 0 {
		querys["marker"] = marker
	}

	querys["limit"] = strconv.Itoa(limit)
	eips := make([]SEipAddress, 0)
	err := DoList(self.ecsClient.Eips.List, querys, &eips)
	for i := range eips {
		eips[i].region = self
	}
	return eips, len(eips), err
}

func (self *SRegion) GetIEips() ([]cloudprovider.ICloudEIP, error) {
	_, err := self.getECSClient()
	if err != nil {
		return nil, err
	}

	marker := ""
	limit := 100
	eips := make([]SEipAddress, 0)
	for {
		var parts []SEipAddress
		parts, count, err := self.GetEips(marker, limit)
		if err != nil {
			return nil, err
		}

		eips = append(eips, parts...)

		if count < limit {
			break
		}

		marker = parts[count-1].ID
	}

	ret := make([]cloudprovider.ICloudEIP, len(eips))
	for i := 0; i < len(eips); i += 1 {
		eips[i].region = self
		ret[i] = &eips[i]
	}
	return ret, nil
}

func (self *SRegion) GetIVpcById(id string) (cloudprovider.ICloudVpc, error) {
	ivpcs, err := self.GetIVpcs()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(ivpcs); i += 1 {
		if ivpcs[i].GetGlobalId() == id {
			return ivpcs[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (self *SRegion) GetIZoneById(id string) (cloudprovider.ICloudZone, error) {
	izones, err := self.GetIZones()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(izones); i += 1 {
		if izones[i].GetGlobalId() == id {
			return izones[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (self *SRegion) GetIEipById(eipId string) (cloudprovider.ICloudEIP, error) {
	eip, err := self.GetEipById(eipId)
	return &eip, err
}

func (self *SRegion) DeleteSecurityGroup(vpcId, secgroupId string) error {
	// todo: implement me
	return nil
}

func (self *SRegion) SyncSecurityGroup(secgroupId string, vpcId string, name string, desc string, rules []secrules.SecurityRule) (string, error) {
	//if len(secgroupId) > 0 {
	//	_, total, err := self.GetSecurityGroups("", []string{secgroupId}, 0, 1)
	//	if err != nil {
	//		return "", err
	//	}
	//	if total == 0 {
	//		secgroupId = ""
	//	}
	//}
	//if len(secgroupId) == 0 {
	//	extID, err := self.CreateSecurityGroup(vpcId, name, desc)
	//	if err != nil {
	//		return "", err
	//	}
	//	secgroupId = extID
	//}
	//return secgroupId, self.syncSecgroupRules(secgroupId, rules)
	// todo: implement me
	return "", nil
}

func (self *SRegion) CreateIVpc(name string, desc string, cidr string) (cloudprovider.ICloudVpc, error) {
	// todo: implement me
	return self.GetIVpcById("")
}

func (self *SRegion) CreateEIP(name string, bwMbps int, chargeType string) (cloudprovider.ICloudEIP, error) {
	// todo: implement me
	var ctype TInternetChargeType
	switch chargeType {
	case models.EIP_CHARGE_TYPE_BY_TRAFFIC:
		ctype = InternetChargeByTraffic
	case models.EIP_CHARGE_TYPE_BY_BANDWIDTH:
		ctype = InternetChargeByBandwidth
	}

	eip, err := self.AllocateEIP(bwMbps, ctype)
	return eip, err
}

func (self *SRegion) GetISnapshots() ([]cloudprovider.ICloudSnapshot, error) {
	snapshots := make([]SSnapshot, 0)
	offset := 0
	limit := 100
	for {
		var parts []SSnapshot
		parts, count, err := self.GetSnapshots("", "", offset, limit)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, parts...)

		if count < limit {
			break
		}

		offset += limit
	}

	ret := make([]cloudprovider.ICloudSnapshot, len(snapshots))
	for i := 0; i < len(snapshots); i += 1 {
		snapshots[i].region = self
		ret[i] = &snapshots[i]
	}
	return ret, nil
}

func (self *SRegion) GetISnapshotById(snapshotId string) (cloudprovider.ICloudSnapshot, error) {
	snapshot, err := self.GetSnapshotById(snapshotId)
	return &snapshot, err
}

func (self *SRegion) GetIHosts() ([]cloudprovider.ICloudHost, error) {
	iHosts := make([]cloudprovider.ICloudHost, 0)

	izones, err := self.GetIZones()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(izones); i += 1 {
		iZoneHost, err := izones[i].GetIHosts()
		if err != nil {
			return nil, err
		}
		iHosts = append(iHosts, iZoneHost...)
	}
	return iHosts, nil
}

func (self *SRegion) GetIHostById(id string) (cloudprovider.ICloudHost, error) {
	izones, err := self.GetIZones()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(izones); i += 1 {
		ihost, err := izones[i].GetIHostById(id)
		if err == nil {
			return ihost, nil
		} else if err != cloudprovider.ErrNotFound {
			return nil, err
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (self *SRegion) GetIStorages() ([]cloudprovider.ICloudStorage, error) {
	iStores := make([]cloudprovider.ICloudStorage, 0)

	izones, err := self.GetIZones()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(izones); i += 1 {
		iZoneStores, err := izones[i].GetIStorages()
		if err != nil {
			return nil, err
		}
		iStores = append(iStores, iZoneStores...)
	}
	return iStores, nil
}

func (self *SRegion) GetIStorageById(id string) (cloudprovider.ICloudStorage, error) {
	izones, err := self.GetIZones()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(izones); i += 1 {
		istore, err := izones[i].GetIStorageById(id)
		if err == nil {
			return istore, nil
		} else if err != cloudprovider.ErrNotFound {
			return nil, err
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (self *SRegion) GetProvider() string {
	return CLOUD_PROVIDER_HUAWEI
}
