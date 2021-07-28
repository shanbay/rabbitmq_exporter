package main

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"time"
)

func init() {
	RegisterExporter("aliveness", newExporterAliveness)
}

var (
	alivenessLabels = []string{"vhost"}

	alivenessGaugeVec = map[string]*prometheus.GaugeVec{
		"vhost.aliveness":              newGaugeVec("aliveness_test", "vhost aliveness test", alivenessLabels),
		"vhost.aliveness.request_time": newGaugeVec("aliveness_test_request_time", "vhost aliveness test request time", alivenessLabels),
	}

	rabbitmqAlivenessMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rabbitmq_aliveness_info",
			Help: "A metric with value 1 status:ok else 0 labeled by aliveness test status, error, reason",
		},
		[]string{"status", "error", "reason"},
	)

	rabbitmqAlivenessReqTimeMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "rabbitmq_aliveness_req_time",
			Help: "A metric with aliveness_req_time by aliveness_test request time",
		},
		[]string{"reqTime"},
	)
)

type exporterAliveness struct {
	alivenessMetrics map[string]*prometheus.GaugeVec
	alivenessInfo    AlivenessInfo
	alivenessReqTime AlivenessReqTime
}

type AlivenessInfo struct {
	Status string
	Error  string
	Reason string
}

type AlivenessReqTime struct {
	ReqTime string
}

func newExporterAliveness() Exporter {
	alivenessGaugeVecActual := alivenessGaugeVec

	if len(config.ExcludeMetrics) > 0 {
		for _, metric := range config.ExcludeMetrics {
			if alivenessGaugeVecActual[metric] != nil {
				delete(alivenessGaugeVecActual, metric)
			}
		}
	}

	return &exporterAliveness{
		alivenessMetrics: alivenessGaugeVecActual,
		alivenessInfo:    AlivenessInfo{},
		alivenessReqTime: AlivenessReqTime{},
	}
}

func (e *exporterAliveness) Collect(ctx context.Context, ch chan<- prometheus.Metric) error {
	t1 := time.Now()
	body, contentType, err := apiRequest(config, "aliveness-test")
	t2 := time.Now()
	var requestTime int64 = t2.Sub(t1).Milliseconds()
	if err != nil {
		return err
	}
	rabbitmqAlivenessReqTimeMetric.Reset()
	e.alivenessReqTime.ReqTime = string(requestTime)
	rabbitmqAlivenessReqTimeMetric.WithLabelValues(e.alivenessReqTime.ReqTime).Set(float64(requestTime))

	reply, err := MakeReply(contentType, body)
	if err != nil {
		return err
	}

	rabbitMqAlivenessData := reply.MakeMap()

	e.alivenessInfo.Status, _ = reply.GetString("status")
	e.alivenessInfo.Error, _ = reply.GetString("error")
	e.alivenessInfo.Reason, _ = reply.GetString("reason")

	rabbitmqAlivenessMetric.Reset()
	var flag float64 = 0
	if e.alivenessInfo.Status == "ok" && requestTime < config.AlivenessReqTime {
		flag = 1
	}
	rabbitmqAlivenessMetric.WithLabelValues(e.alivenessInfo.Status, e.alivenessInfo.Error, e.alivenessInfo.Reason).Set(flag)

	//log.WithField("alivenesswData", rabbitMqAlivenessData).Debug("Aliveness data")
	log.WithField("alivenesswData", e.alivenessInfo).Debug("Aliveness data")
	for key, gauge := range e.alivenessMetrics {
		if value, ok := rabbitMqAlivenessData[key]; ok {
			log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set aliveness metric for key")
			gauge.WithLabelValues(e.alivenessInfo.Status).Set(value)
		}
	}

	if ch != nil {
		rabbitmqAlivenessMetric.Collect(ch)
		rabbitmqAlivenessReqTimeMetric.Collect(ch)
		for _, gauge := range e.alivenessMetrics {
			gauge.Collect(ch)
		}
	}
	return nil
}

func (e exporterAliveness) Describe(ch chan<- *prometheus.Desc) {
	rabbitmqVersionMetric.Describe(ch)
	rabbitmqAlivenessReqTimeMetric.Describe(ch)
	for _, gauge := range e.alivenessMetrics {
		gauge.Describe(ch)
	}
}
