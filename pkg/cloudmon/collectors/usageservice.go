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

package collectors

import (
	"fmt"
	"strings"
	"time"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	"yunion.io/x/onecloud/pkg/cloudmon/collectors/common"
	"yunion.io/x/onecloud/pkg/cloudmon/options"
	"yunion.io/x/onecloud/pkg/mcclient"
	modules "yunion.io/x/onecloud/pkg/mcclient/modules/compute"
	"yunion.io/x/onecloud/pkg/mcclient/modules/image"
	"yunion.io/x/onecloud/pkg/util/influxdb"
	"yunion.io/x/onecloud/pkg/util/shellutils"
)

var config map[string]string = map[string]string{
	"hypervisors":           "host-type",
	"compute_engine_brands": "provider",
}
var measureMent string = "usage"

func init() {
	shellutils.R(&options.ReportOptions{}, "report-usage", "Report Usage", reportUsage)
	factory := UsageColectorFactory{}
	common.RegisterFactory(&factory)
}

type UsageColectorFactory struct {
	common.CommonReportFactory
}

func (p UsageColectorFactory) NewCloudReport(provider *common.SProvider, session *mcclient.ClientSession, args *options.ReportOptions, operatorType string) common.ICloudReport {
	return &SUsageColectorReport{
		common.CloudReportBase{
			SProvider: nil,
			Session:   session,
			Args:      args,
			Operator:  string(common.USAGE),
		},
	}
}

func (p UsageColectorFactory) GetId() string {
	return string(common.USAGE)
}

type SUsageColectorReport struct {
	common.CloudReportBase
}

func (self *SUsageColectorReport) Report() error {
	err := reportUsage(self.Session, self.Args)
	if err != nil {
		return errors.Wrap(err, "usageCollector err")
	}
	return nil
}

func reportUsage(session *mcclient.ClientSession, args *options.ReportOptions) error {
	dataList := make([]influxdb.SMetricData, 0)
	nowTime := time.Now()
	//?????????????????????
	imageUsageFields, err := getImageUsageFields(session)
	if err != nil {
		return err
	}
	//????????????Usage?????????????????????metric???
	imageUsageFieldsDict := imageUsageFields.(*jsonutils.JSONDict)
	capabilitesQuery := jsonutils.NewDict()
	capabilitesQuery.Add(jsonutils.NewString("system"), "scope")
	capabilites, err := modules.Capabilities.List(session, capabilitesQuery)
	if err != nil {
		return err
	}
	//??????capabilities??????????????????hypevisors???brands
	for i := 0; i < len(capabilites.Data); i++ {
		capabilitesObj := capabilites.Data[i]
		capDict, ok := capabilitesObj.(*jsonutils.JSONDict)
		if !ok {
			return errors.ErrClient
		}
		for _, capKey := range capDict.SortedKeys() {
			if _, ok := config[capKey]; ok {
				hypeOrBrandObj, _ := capDict.Get(capKey)
				if hypeOrBrandObj != nil {
					hypeOrBrandArr, _ := hypeOrBrandObj.(*jsonutils.JSONArray)
					for i := 0; i < len(hypeOrBrandArr.Value()); i++ {
						hypeOrBrand := hypeOrBrandArr.Value()[i].(*jsonutils.JSONString)
						dataList, err = packMetricList(session, dataList, imageUsageFieldsDict, config[capKey],
							hypeOrBrand.String(), nowTime)
						if err != nil {
							return err
						}
					}
				}
			}
		}
	}
	//??????host-type==""??????????????????onecloud-????????????-?????? ??????????????????
	dataList, err = packMetricList(session, dataList, imageUsageFieldsDict, "host-type", "", nowTime)
	//?????????/?????? ?????????usage
	data, err := getDomainAndProjectServerUsage(session, nowTime)
	if err != nil {
		return errors.Wrap(err, "getDomainAndProjectServerUsage err")
	}
	dataList = append(dataList, data...)
	//??????influDb
	return sendMetrics(session, dataList, args.Debug)
}

//??????capabilities??????hypevisors???brands???????????????????????????Metric
func packMetricList(session *mcclient.ClientSession, dataList []influxdb.SMetricData,
	imageUsageFieldsDict *jsonutils.JSONDict, paramKey string,
	paramValue string, nowTime time.Time) (rtnList []influxdb.SMetricData, err error) {
	//query???sql?????????????????????compute???????????????
	query := jsonutils.NewDict()
	query.Add(jsonutils.NewString("system"), "scope")
	if paramValue != "" {
		if paramKey == "host-type" {
			if paramValue == "kvm" {
				query.Add(jsonutils.NewString("hypervisor"), paramKey)
			}
			query.Add(jsonutils.NewString(paramValue), paramKey)
		}
		query.Add(jsonutils.NewString(paramValue), paramKey)
	}
	metric := &influxdb.SMetricData{Name: measureMent, Timestamp: nowTime}
	//???????????????????????????????????????SMetricData??????metric
	metric, _ = jsonTometricData(imageUsageFieldsDict, metric, "metric")
	//compute???????????????
	respObj, err := modules.Usages.GetGeneralUsage(session, query)
	if err != nil {
		return nil, err
	}
	respObjDict := respObj.(*jsonutils.JSONDict)
	//???????????????????????????????????????SMetricData??????metric
	metric, _ = jsonTometricData(respObjDict, metric, "metric")
	if paramValue != "" {
		metric.Tags = append(metric.Tags, influxdb.SKeyValue{
			Key: paramKey, Value: paramValue,
		})
	} else {
		metric.Tags = append(metric.Tags, influxdb.SKeyValue{
			Key: paramKey, Value: "all",
		})
	}
	dataList = append(dataList, *metric)
	return dataList, nil
}

//?????????????????????
func getImageUsageFields(session *mcclient.ClientSession) (jsonutils.JSONObject, error) {
	respObj, e := (&image.ImageUsages).GetUsage(session, nil)
	if e != nil {
		return nil, e
	}
	respDict, ok := respObj.(*jsonutils.JSONDict)
	if !ok {
		return nil, jsonutils.ErrInvalidJsonDict
	}
	return respDict, nil
}

//???JSONDict??????????????????SMetricData???
func jsonTometricData(obj *jsonutils.JSONDict, metric *influxdb.SMetricData,
	metricDataType string) (*influxdb.SMetricData, error) {

	objMap, err := obj.GetMap()
	if err != nil {
		return nil, errors.Wrap(err, "obj.GetMap")
	}
	tagPairs := make([]influxdb.SKeyValue, 0)
	metricPairs := make([]influxdb.SKeyValue, 0)
	for k, v := range objMap {
		val, _ := v.GetString()
		if metricDataType == "tag" {
			tagPairs = append(tagPairs, influxdb.SKeyValue{
				Key: k, Value: val,
			})
		} else if metricDataType == "metric" {
			metricPairs = append(metricPairs, influxdb.SKeyValue{
				Key: k, Value: val,
			})
		}
	}
	metric.Tags = append(metric.Tags, tagPairs...)
	metric.Metrics = append(metric.Metrics, metricPairs...)
	return metric, nil
}

type resourceUsage struct {
	Count string
	Id    string
	Name  string
}

func getDomainAndProjectServerUsage(session *mcclient.ClientSession, nowTime time.Time) ([]influxdb.SMetricData, error) {
	metrics := make([]influxdb.SMetricData, 0)
	param := jsonutils.NewDict()
	param.Set("scope", jsonutils.NewString("system"))
	for urlKey, tag := range map[string]string{
		"domain-statistics":  "domain_id.project_domain",
		"project-statistics": "tenant_id.tenant",
	} {
		jsonObject, err := modules.Servers.GetById(session, urlKey, param)
		if err != nil {
			return nil, errors.Wrapf(err, "get server-%s err", urlKey)
		}
		usageArr := jsonObject.(*jsonutils.JSONArray)
		for i := 0; i < usageArr.Size(); i++ {
			metric := influxdb.SMetricData{Name: measureMent, Timestamp: nowTime}
			usageObj, _ := usageArr.GetAt(i)
			usage := new(resourceUsage)
			usageObj.Unmarshal(usage)
			key := ""
			if strings.Contains(urlKey, "domain") {
				key = "domain"
			} else {
				key = "project"
			}
			metric.Metrics = append(metric.Metrics, influxdb.SKeyValue{
				Key:   fmt.Sprintf("%s.servers", key),
				Value: usage.Count,
			})
			tagArr := strings.Split(tag, ".")
			metric.Tags = append(metric.Tags, influxdb.SKeyValue{
				Key:   tagArr[0],
				Value: usage.Id,
			}, influxdb.SKeyValue{
				Key:   tagArr[1],
				Value: usage.Name,
			})
			metrics = append(metrics, metric)
		}
	}
	return metrics, nil
}
