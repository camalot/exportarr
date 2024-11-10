package collector

import (
	"strconv"

	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type indexerCollector struct {
	config        *config.ArrConfig // App configuration
	indexerMetric *prometheus.Desc  // Total number of root folders
	errorMetric   *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewIndexerCollector(c *config.ArrConfig) *indexerCollector {
	return &indexerCollector{
		config: c,
		indexerMetric: prometheus.NewDesc(
			prometheus.BuildFQName(c.App, "indexer", ""),
			"Indexer Metrics. 0 = Disabled, RssEnabled = 1 | AutoSearchEnabled = 2 | InteractiveSearchEnabled = 3",
			[]string{"protocol", "name", "priority", "implementation"},
			prometheus.Labels{"url": c.URL},
		),
		errorMetric: prometheus.NewDesc(
			prometheus.BuildFQName(c.App, "indexer", "collector_error"),
			"Error while collecting metrics",
			nil,
			prometheus.Labels{"url": c.URL},
		),
	}
}

func (collector *indexerCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.indexerMetric
}

func (collector *indexerCollector) Collect(ch chan<- prometheus.Metric) {
	log := zap.S().With("collector", "indexer")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
	indexers := model.ArrIndexer{}
	if err := c.DoRequest("indexer", &indexers); err != nil {
		log.Errorw("Error getting indexer",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	if len(indexers) > 0 {
		for _, indexer := range indexers {
			status := 0
			if indexer.EnableRss {
				status &= 1
			}
			if indexer.EnableAutomaticSearch {
				status &= 2
			}
			if indexer.EnableInteractiveSearch {
				status &= 3
			}
			ch <- prometheus.MustNewConstMetric(
				collector.indexerMetric,
				prometheus.GaugeValue,
				float64(status),
				indexer.Protocol,
				indexer.Name,
				strconv.Itoa(indexer.Priority),
				indexer.Implementation,
			)
		}
	}
}
