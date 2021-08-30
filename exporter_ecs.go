package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	RegisterExporter("ecs", newExporterEcsRabbitmq)
}

type AliEcsMonitorResponse struct {
	RequestID   string `json:"RequestId"`
	MonitorData struct {
		InstanceMonitorData []struct {
			IOPSRead          int       `json:"IOPSRead"`
			IntranetBandwidth int       `json:"IntranetBandwidth"`
			IOPSWrite         int       `json:"IOPSWrite"`
			InstanceID        string    `json:"InstanceId"`
			IntranetTX        int       `json:"IntranetTX"`
			CPU               int       `json:"CPU"`
			BPSRead           int       `json:"BPSRead"`
			IntranetRX        int       `json:"IntranetRX"`
			TimeStamp         time.Time `json:"TimeStamp"`
			InternetBandwidth int       `json:"InternetBandwidth"`
			InternetTX        int       `json:"InternetTX"`
			InternetRX        int       `json:"InternetRX"`
			BPSWrite          int       `json:"BPSWrite"`
		} `json:"InstanceMonitorData"`
	} `json:"MonitorData"`
}

type ListAliEcsResponse struct {
	Instances struct {
		Instance []struct {
			ResourceGroupID     string `json:"ResourceGroupId"`
			Memory              int    `json:"Memory"`
			InstanceChargeType  string `json:"InstanceChargeType"`
			CPU                 int    `json:"Cpu"`
			OSName              string `json:"OSName"`
			InstanceNetworkType string `json:"InstanceNetworkType"`
			InnerIpddress       struct {
				IPAddress []interface{} `json:"IpAddress"`
			} `json:"InnerIpddress,omitempty"`
			ExpiredTime string `json:"ExpiredTime"`
			ImageID     string `json:"ImageId"`
			EipAddress  struct {
				AllocationID       string `json:"AllocationId"`
				IPAddress          string `json:"IpAddress"`
				InternetChargeType string `json:"InternetChargeType"`
			} `json:"EipAddress"`
			HostName string `json:"HostName"`
			Tags     struct {
				Tag []struct {
					TagKey   string `json:"TagKey"`
					TagValue string `json:"TagValue"`
				} `json:"Tag"`
			} `json:"Tags"`
			VlanID             string `json:"VlanId"`
			Status             string `json:"Status"`
			HibernationOptions struct {
				Configured bool `json:"Configured"`
			} `json:"HibernationOptions"`
			MetadataOptions struct {
				HTTPTokens   string `json:"HttpTokens"`
				HTTPEndpoint string `json:"HttpEndpoint"`
			} `json:"MetadataOptions"`
			InstanceID  string `json:"InstanceId"`
			StoppedMode string `json:"StoppedMode"`
			CPUOptions  struct {
				ThreadsPerCore int    `json:"ThreadsPerCore"`
				Numa           string `json:"Numa"`
				CoreCount      int    `json:"CoreCount"`
			} `json:"CpuOptions"`
			StartTime          string `json:"StartTime"`
			DeletionProtection bool   `json:"DeletionProtection"`
			SecurityGroupIds   struct {
				SecurityGroupID []string `json:"SecurityGroupId"`
			} `json:"SecurityGroupIds"`
			VpcAttributes struct {
				PrivateIPAddress struct {
					IPAddress []string `json:"IpAddress"`
				} `json:"PrivateIpAddress"`
				VpcID        string `json:"VpcId"`
				VSwitchID    string `json:"VSwitchId"`
				NatIPAddress string `json:"NatIpAddress"`
			} `json:"VpcAttributes"`
			InternetChargeType         string `json:"InternetChargeType"`
			InstanceName               string `json:"InstanceName"`
			DeploymentSetID            string `json:"DeploymentSetId"`
			InternetMaxBandwidthOut    int    `json:"InternetMaxBandwidthOut"`
			SerialNumber               string `json:"SerialNumber"`
			OSType                     string `json:"OSType"`
			CreationTime               string `json:"CreationTime"`
			AutoReleaseTime            string `json:"AutoReleaseTime"`
			Description                string `json:"Description"`
			InstanceTypeFamily         string `json:"InstanceTypeFamily"`
			DedicatedInstanceAttribute struct {
				Tenancy  string `json:"Tenancy"`
				Affinity string `json:"Affinity"`
			} `json:"DedicatedInstanceAttribute"`
			PublicIPAddress struct {
				IPAddress []interface{} `json:"IpAddress"`
			} `json:"PublicIpAddress"`
			GPUSpec           string `json:"GPUSpec"`
			NetworkInterfaces struct {
				NetworkInterface []struct {
					Type               string `json:"Type"`
					PrimaryIPAddress   string `json:"PrimaryIpAddress"`
					MacAddress         string `json:"MacAddress"`
					NetworkInterfaceID string `json:"NetworkInterfaceId"`
					PrivateIPSets      struct {
						PrivateIPSet []struct {
							PrivateIPAddress string `json:"PrivateIpAddress"`
							Primary          bool   `json:"Primary"`
						} `json:"PrivateIpSet"`
					} `json:"PrivateIpSets"`
				} `json:"NetworkInterface"`
			} `json:"NetworkInterfaces"`
			SpotPriceLimit             float64 `json:"SpotPriceLimit"`
			DeviceAvailable            bool    `json:"DeviceAvailable"`
			SaleCycle                  string  `json:"SaleCycle"`
			InstanceType               string  `json:"InstanceType"`
			SpotStrategy               string  `json:"SpotStrategy"`
			OSNameEn                   string  `json:"OSNameEn"`
			IoOptimized                bool    `json:"IoOptimized"`
			ZoneID                     string  `json:"ZoneId"`
			ClusterID                  string  `json:"ClusterId"`
			EcsCapacityReservationAttr struct {
				CapacityReservationPreference string `json:"CapacityReservationPreference"`
				CapacityReservationID         string `json:"CapacityReservationId"`
			} `json:"EcsCapacityReservationAttr"`
			DedicatedHostAttribute struct {
				DedicatedHostID        string `json:"DedicatedHostId"`
				DedicatedHostName      string `json:"DedicatedHostName"`
				DedicatedHostClusterID string `json:"DedicatedHostClusterId"`
			} `json:"DedicatedHostAttribute"`
			GPUAmount      int `json:"GPUAmount"`
			OperationLocks struct {
				LockReason []interface{} `json:"LockReason"`
			} `json:"OperationLocks"`
			InternetMaxBandwidthIn int    `json:"InternetMaxBandwidthIn"`
			Recyclable             bool   `json:"Recyclable"`
			RegionID               string `json:"RegionId"`
			CreditSpecification    string `json:"CreditSpecification"`
			InnerIPAddress         struct {
				IPAddress []interface{} `json:"IpAddress"`
			} `json:"InnerIpAddress,omitempty"`
		} `json:"Instance"`
	} `json:"Instances"`
	TotalCount int    `json:"TotalCount"`
	NextToken  string `json:"NextToken"`
	PageSize   int    `json:"PageSize"`
	RequestID  string `json:"RequestId"`
	PageNumber int    `json:"PageNumber"`
}

var (
	EcsRabbitmqLabels = []string{"aliecs"}

	ecsRabbitmqGaugeVec = map[string]*prometheus.GaugeVec{
		"ecs.rabbitmq.cpu": newGaugeVec("ecs_cpu", "ecs rabbit cpu", EcsRabbitmqLabels),
	}

	ecsRabbitmqMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ecs_rabbitmq_cpu",
			Help: "A metric ecs rabbitmq labeled by instanceName, cpu",
		},
		[]string{"instanceName", "cpu"},
	)
)

type exporterEcsRabbitmq struct {
	ecsRabbitmqMetrics map[string]*prometheus.GaugeVec
	ecsRabbitmqInfo    EcsRabbitmqInfo
}

type EcsRabbitmqInfo struct {
	instanceId        string
	instanceName      string
	cpu               int
	iOPSRead          int
	intranetBandwidth int
	iOPSWrite         int
	intranetTX        int
	bPSRead           int
	intranetRX        int
	timeStamp         time.Time
	internetBandwidth int
	internetTX        int
	internetRX        int
	bPSWrite          int
}

func newExporterEcsRabbitmq() Exporter {
	ecsRabbitmqGaugeVecActual := ecsRabbitmqGaugeVec

	if len(config.ExcludeMetrics) > 0 {
		for _, metric := range config.ExcludeMetrics {
			if ecsRabbitmqGaugeVecActual[metric] != nil {
				delete(ecsRabbitmqGaugeVecActual, metric)
			}
		}
	}

	return &exporterEcsRabbitmq{
		ecsRabbitmqMetrics: ecsRabbitmqGaugeVecActual,
		ecsRabbitmqInfo:    EcsRabbitmqInfo{},
	}
}

func (e *exporterEcsRabbitmq) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	aliEcsRabbitInfos, err := aliEcsRequest(config)
	if err != nil {
		return err
	}

	ecsRabbitmqMetric.Reset()
	for _, info := range aliEcsRabbitInfos {
		e.ecsRabbitmqInfo = info
		ecsRabbitmqMetric.WithLabelValues(e.ecsRabbitmqInfo.instanceName, "CPU").Set(float64(e.ecsRabbitmqInfo.cpu))
	}

	if ch != nil {
		ecsRabbitmqMetric.Collect(ch)
		for _, gauge := range e.ecsRabbitmqMetrics {
			gauge.Collect(ch)
		}
	}
	return nil
}

func aliEcsRequest(config rabbitExporterConfig) ([]EcsRabbitmqInfo, error) {
	accessKey := config.AliAccessKey
	accessSecret := config.AliAccessSecret
	regionId := config.RegionId
	client, err := ecs.NewClientWithAccessKey(regionId, accessKey, accessSecret)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error while create con to ali")
		return nil, errors.New("new client with ali ecs is error")
	}

	requestEcsList := ecs.CreateDescribeInstancesRequest()
	requestEcsList.Scheme = "https"
	requestEcsList.InstanceName = "rabbit*"

	response, err := client.DescribeInstances(requestEcsList)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Error while DescribeInstances request")
		return nil, errors.New("error while DescribeInstances request")
	}

	var listAliEcsResp ListAliEcsResponse
	err = json.Unmarshal(response.GetHttpContentBytes(), &listAliEcsResp)
	instances := listAliEcsResp.Instances.Instance

	if len(instances) < 3 {
		log.WithFields(log.Fields{"error": err}).Error("Error rabbitmq instance count")
		return nil, errors.New("error rabbitmq instance count")
	}

	requestEcsMonitor := ecs.CreateDescribeInstanceMonitorDataRequest()
	requestEcsMonitor.Scheme = "https"

	var ecsRabbitmqInfos []EcsRabbitmqInfo

	for _, instance := range instances {
		instanceId := instance.InstanceID
		instanceName := instance.InstanceName

		requestEcsMonitor.InstanceId = instanceId
		// 该接口只支持粒度为分钟
		requestEcsMonitor.StartTime = time.Now().Add(-2 * time.Minute).UTC().Format(time.RFC3339)
		requestEcsMonitor.EndTime = time.Now().UTC().Format(time.RFC3339)
		requestEcsMonitor.Period = requests.NewInteger(60)
		response, err := client.DescribeInstanceMonitorData(requestEcsMonitor)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Error API DescribeInstanceMonitorData")
			return nil, errors.New("error API DescribeInstanceMonitorData")
		}

		var aliEcsMonitorResp AliEcsMonitorResponse
		err = json.Unmarshal(response.GetHttpContentBytes(), &aliEcsMonitorResp)
		l := len(aliEcsMonitorResp.MonitorData.InstanceMonitorData)
		if l == 0 {
			log.WithFields(log.Fields{"error": err}).Error("Error get ali ecs monitor data")
			return nil, errors.New("error get ali ecs monitor data")
		}

		var ecsRabbitmqInfo EcsRabbitmqInfo
		ecsRabbitmqInfo.instanceName = instanceName
		ecsRabbitmqInfo.instanceId = instanceId
		ecsRabbitmqInfo.iOPSRead = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].IOPSRead
		ecsRabbitmqInfo.intranetBandwidth = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].IntranetBandwidth
		ecsRabbitmqInfo.iOPSWrite = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].IOPSWrite
		ecsRabbitmqInfo.intranetTX = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].IntranetTX
		ecsRabbitmqInfo.cpu = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].CPU
		ecsRabbitmqInfo.bPSRead = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].BPSRead
		ecsRabbitmqInfo.intranetRX = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].IntranetRX
		ecsRabbitmqInfo.timeStamp = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].TimeStamp
		ecsRabbitmqInfo.internetBandwidth = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].InternetBandwidth
		ecsRabbitmqInfo.internetTX = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].InternetTX
		ecsRabbitmqInfo.internetRX = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].InternetRX
		ecsRabbitmqInfo.bPSWrite = aliEcsMonitorResp.MonitorData.InstanceMonitorData[l-1].BPSWrite

		ecsRabbitmqInfos = append(ecsRabbitmqInfos, ecsRabbitmqInfo)
	}
	return ecsRabbitmqInfos, err
}

func (e exporterEcsRabbitmq) Describe(ch chan<- *prometheus.Desc) {
	ecsRabbitmqMetric.Describe(ch)
	for _, gauge := range e.ecsRabbitmqMetrics {
		gauge.Describe(ch)
	}
}
