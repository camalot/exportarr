package collector

import (
	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type blocklistCollector struct {
	config       *config.ArrConfig // App configuration
	blockedMetric *prometheus.Desc  //
	errorMetric  *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewBlocklistCollector(conf *config.ArrConfig) *blocklistCollector {
	return &blocklistCollector{
		config: conf,
		blockedMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "blocklist", "total"),
			"Number of blocked items",
			[]string{},
			prometheus.Labels{"url": conf.URL},
		),

		errorMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "indexer", "collector_error"),
			"Error while collecting metrics",
			nil,
			prometheus.Labels{"url": conf.URL},
		),
	}
}

func (collector *blocklistCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.blockedMetric
}

func (collector *blocklistCollector) Collect(ch chan<- prometheus.Metric) {
	log := zap.S().With("collector", "blocklist")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	blocklist := model.BlockList{}
	params := client.QueryParams{}
	params.Add("page", "1")

	if err := c.DoRequest("blocklist", &blocklist, params); err != nil {
		log.Errorw("Error getting blocklist",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		collector.blockedMetric,
		prometheus.GaugeValue,
		float64(blocklist.TotalRecords),
	)

}
