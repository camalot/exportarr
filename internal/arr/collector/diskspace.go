package collector

import (
	"github.com/onedr0p/exportarr/internal/arr/client"
	"github.com/onedr0p/exportarr/internal/arr/config"
	"github.com/onedr0p/exportarr/internal/arr/model"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type diskspaceCollector struct {
	config           *config.ArrConfig // App configuration
	freeSpaceMetric  *prometheus.Desc  //
	totalSpaceMetric *prometheus.Desc  //
	errorMetric      *prometheus.Desc  // Error Description for use with InvalidMetric
}

func NewDiskSpaceCollector(conf *config.ArrConfig) *diskspaceCollector {
	return &diskspaceCollector{
		config: conf,
		freeSpaceMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "diskspace", "free_bytes_total"),
			"Freespace in bytes",
			[]string{"path", "label"},
			prometheus.Labels{"url": conf.URL},
		),
		totalSpaceMetric: prometheus.NewDesc(
			prometheus.BuildFQName(conf.App, "diskspace", "bytes_total"),
			"total space in bytes",
			[]string{"path", "label"},
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

func (collector *diskspaceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.freeSpaceMetric
	ch <- collector.totalSpaceMetric
}

func (collector *diskspaceCollector) Collect(ch chan<- prometheus.Metric) {
	log := zap.S().With("collector", "diskspace")
	c, err := client.NewClient(collector.config)
	if err != nil {
		log.Errorw("Error creating client",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}
	disks := model.DiskSpace{}
	if err := c.DoRequest("diskspace", &disks); err != nil {
		log.Errorw("Error getting diskspace",
			"error", err)
		ch <- prometheus.NewInvalidMetric(collector.errorMetric, err)
		return
	}

	if len(disks) > 0 {
		for _, disk := range disks {
			ch <- prometheus.MustNewConstMetric(
				collector.freeSpaceMetric,
				prometheus.GaugeValue,
				float64(disk.FreeSpace),
				disk.Path,
				disk.Label,
			)

			ch <- prometheus.MustNewConstMetric(
				collector.totalSpaceMetric,
				prometheus.GaugeValue,
				float64(disk.TotalSpace),
				disk.Path,
				disk.Label,
			)
		}
	}
}
