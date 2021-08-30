package main

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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
	var flag float64 = 0
	t1 := time.Now()
	body, contentType, err := apiRequest(config, "aliveness-test")
	t2 := time.Now()
	rabbitmqAlivenessMetric.Reset()
	if err != nil {
		// 请求接口出错
		log.Error("aliveness-test api req fail")
		rabbitmqAlivenessMetric.WithLabelValues(e.alivenessInfo.Status, e.alivenessInfo.Error, e.alivenessInfo.Reason).Set(flag)
		guageCollect(e, ch)
		return nil
	}

	reply, err := MakeReply(contentType, body)
	if err != nil {
		// 解析响应体出错
		log.Error("aliveness-test api parse fail")
		rabbitmqAlivenessMetric.WithLabelValues(e.alivenessInfo.Status, e.alivenessInfo.Error, e.alivenessInfo.Reason).Set(flag)
		guageCollect(e, ch)
		return nil
	}

	var requestTime int64 = t2.Sub(t1).Milliseconds()
	rabbitmqAlivenessReqTimeMetric.Reset()
	// e.alivenessReqTime.ReqTime = strconv.FormatInt(requestTime,10)

	rabbitmqAlivenessReqTimeMetric.WithLabelValues("").Set(float64(requestTime))

	rabbitMqAlivenessData := reply.MakeMap()

	e.alivenessInfo.Status, _ = reply.GetString("status")
	e.alivenessInfo.Error, _ = reply.GetString("error")
	e.alivenessInfo.Reason, _ = reply.GetString("reason")

	if e.alivenessInfo.Status == "ok" && requestTime < config.AlivenessReqTime {
		flag = 1
	}
	rabbitmqAlivenessMetric.WithLabelValues(e.alivenessInfo.Status, e.alivenessInfo.Error, e.alivenessInfo.Reason).Set(flag)

	// log.WithField("alivenesswData", rabbitMqAlivenessData).Debug("Aliveness data")
	log.WithField("alivenesswData", e.alivenessInfo).Debug("Aliveness data")
	for key, gauge := range e.alivenessMetrics {
		if value, ok := rabbitMqAlivenessData[key]; ok {
			log.WithFields(log.Fields{"key": key, "value": value}).Debug("Set aliveness metric for key")
			gauge.WithLabelValues(e.alivenessInfo.Status).Set(value)
		}
	}

	guageCollect(e, ch)
	return nil
}

func guageCollect(e *exporterAliveness, ch chan<- prometheus.Metric) {
	if ch != nil {
		rabbitmqAlivenessMetric.Collect(ch)
		rabbitmqAlivenessReqTimeMetric.Collect(ch)
		for _, gauge := range e.alivenessMetrics {
			gauge.Collect(ch)
		}
	}
}

func (e exporterAliveness) Describe(ch chan<- *prometheus.Desc) {
	rabbitmqAlivenessMetric.Describe(ch)
	rabbitmqAlivenessReqTimeMetric.Describe(ch)
	for _, gauge := range e.alivenessMetrics {
		gauge.Describe(ch)
	}
}
