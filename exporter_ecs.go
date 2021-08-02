package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)


func init() {
	RegisterExporter("ecs", newExporterEcsRabbitmq)
}


type AliEcsResponse struct {
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

var (
	EcsRabbitmqLabels = []string{"aliecs"}

	ecsRabbitmqGaugeVec = map[string]*prometheus.GaugeVec{
		"ecs.rabbitmq.cpu": newGaugeVec("ecs_cpu", "ecs rabbit cpu", EcsRabbitmqLabels),
	}

	ecsRabbitmqMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ecs_rabbitmq_cpu",
			Help: "A metric ecs rabbitmq labeled by instanceId, cpu",
		},
		[]string{"instanceId", "cpu"},
	)
)

type exporterEcsRabbitmq struct {
	ecsRabbitmqMetrics map[string]*prometheus.GaugeVec
	ecsRabbitmqInfo    EcsRabbitmqInfo
}

type EcsRabbitmqInfo struct {
	instanceId        string
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
		ecsRabbitmqMetric.WithLabelValues(e.ecsRabbitmqInfo.instanceId, "CPU").Set(float64(e.ecsRabbitmqInfo.cpu))
	}

	//for key, gauge := range e.ecsRabbitmqMetrics {
	//	if value, ok := rabbitMqAlivenessData[key]; ok {
	//		log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set aliveness metric for key")
	//		gauge.WithLabelValues(e.alivenessInfo.Status).Set(value)
	//	}
	//}

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
	request := ecs.CreateDescribeInstanceMonitorDataRequest()
	request.Scheme = "https"
	
	var ecsRabbitmqInfos []EcsRabbitmqInfo

	instanceIds := config.InstanceIds
	for _, instanceId := range instanceIds {
		request.InstanceId = instanceId
		// 该接口只支持粒度为分钟
		request.StartTime = time.Now().Add(-2 * time.Minute).UTC().Format(time.RFC3339)
		request.EndTime = time.Now().UTC().Format(time.RFC3339)
		request.Period = requests.NewInteger(60)
		response, err := client.DescribeInstanceMonitorData(request)
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Error("Error API DescribeInstanceMonitorData")
			return nil, errors.New("Error API DescribeInstanceMonitorData")
		}

		var aliEcsResp AliEcsResponse
		err = json.Unmarshal(response.GetHttpContentBytes(), &aliEcsResp)
		l := len(aliEcsResp.MonitorData.InstanceMonitorData)
		if l == 0{
			log.WithFields(log.Fields{"error": err}).Error("Error get ali ecs monitor data")
			return nil, errors.New("Error get ali ecs monitor data")
		}

		var ecsRabbitmqInfo EcsRabbitmqInfo
		ecsRabbitmqInfo.iOPSRead = aliEcsResp.MonitorData.InstanceMonitorData[l-1].IOPSRead
		ecsRabbitmqInfo.intranetBandwidth = aliEcsResp.MonitorData.InstanceMonitorData[l-1].IntranetBandwidth
		ecsRabbitmqInfo.iOPSWrite = aliEcsResp.MonitorData.InstanceMonitorData[l-1].IOPSWrite
		ecsRabbitmqInfo.instanceId = aliEcsResp.MonitorData.InstanceMonitorData[l-1].InstanceID
		ecsRabbitmqInfo.intranetTX = aliEcsResp.MonitorData.InstanceMonitorData[l-1].IntranetTX
		ecsRabbitmqInfo.cpu = aliEcsResp.MonitorData.InstanceMonitorData[l-1].CPU
		ecsRabbitmqInfo.bPSRead = aliEcsResp.MonitorData.InstanceMonitorData[l-1].BPSRead
		ecsRabbitmqInfo.intranetRX = aliEcsResp.MonitorData.InstanceMonitorData[l-1].IntranetRX
		ecsRabbitmqInfo.timeStamp = aliEcsResp.MonitorData.InstanceMonitorData[l-1].TimeStamp
		ecsRabbitmqInfo.internetBandwidth = aliEcsResp.MonitorData.InstanceMonitorData[l-1].InternetBandwidth
		ecsRabbitmqInfo.internetTX = aliEcsResp.MonitorData.InstanceMonitorData[l-1].InternetTX
		ecsRabbitmqInfo.internetRX = aliEcsResp.MonitorData.InstanceMonitorData[l-1].InternetRX
		ecsRabbitmqInfo.bPSWrite = aliEcsResp.MonitorData.InstanceMonitorData[l-1].BPSWrite

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
