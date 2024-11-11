package collector

import (
	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type logsCollector struct {
	config      *config.ArrConfig // App configuration
	logsMetric  *prometheus.Desc  //
	errorMetric *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewLogsCollector(conf *config.ArrConfig) *logsCollector {
	return &logsCollector{
		config: conf,
		logsMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "logs", "total"),
			"Total number of logs",
			[]string{"level"},
			prometheus.Labels{"url": conf.URL},
		),
		errorMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "logs", "collector_error"),
			"Error while collecting metrics",
			nil,
			prometheus.Labels{"url": conf.URL},
		),
	}
}

func (collector *logsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.logsMetric
}

func (collector *logsCollector) Collect(ch chan<- prometheus.Metric) {
	log := zap.S().With("collector", "logs")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
	levels := []string{"info", "warn", "error", "fatal"}
	logs := model.Logs{}

	// loop through each log level
	for _, level := range levels {
			params := client.QueryParams{}
			params.Add("page", "1")
			params.Add("pageSize", "1")
			params.Add("sortDirection", "descending")
			params.Add("level", level)

		if err := c.DoRequest("logs", &logs); err != nil {
			log.Errorw("Error getting logs",
				"error", err)
			ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
			return
		}

		ch <- prometheus.MustNewConstMetric(
			collector.logsMetric,
			prometheus.GaugeValue,
			float64(logs.TotalRecords),
			level,
		)
	}
}
